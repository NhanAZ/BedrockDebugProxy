<?php

declare(strict_types=1);

namespace BedrockDebugProxy\GolemAgent;

use Golem\Golem;
use pocketmine\entity\Entity;
use pocketmine\item\Item;
use pocketmine\item\StringToItemParser;
use pocketmine\math\Vector3;
use pocketmine\player\GameMode;

final class Actions{
	public const METHODS = [
		'chat' => ['message'], 'command' => ['command'], 'op' => [], 'deop' => [],
		'grant' => ['permission', 'value'], 'deny' => ['permission'], 'gamemode' => ['mode'],
		'teleport' => ['x', 'y', 'z'], 'walkTo' => ['x', 'y', 'z', 'speed'], 'walk' => ['x', 'z', 'speed'],
		'jump' => [], 'sneak' => ['value'], 'sprint' => ['value'], 'look' => ['yaw', 'pitch'],
		'give' => ['item', 'count'], 'hold' => ['item', 'count'], 'inventory' => [], 'setSlot' => ['slot', 'item', 'count'],
		'breakBlock' => ['x', 'y', 'z'], 'interactBlock' => ['x', 'y', 'z', 'face'], 'useItem' => [],
		'attack' => ['target'], 'interactEntity' => ['target'], 'respawn' => [], 'quit' => ['reason'],
		'health' => ['value'], 'hunger' => ['value'], 'kill' => [],
		'clickSlot' => ['slot'], 'closeWindow' => [], 'window' => [], 'waitForWindow' => ['ticks'],
		'formData' => [], 'submitForm' => ['data'], 'clickButton' => ['button'], 'closeForm' => [],
		'messages' => [], 'rawMessages' => [], 'lastMessage' => [], 'titles' => [], 'subtitles' => [],
		'actionBars' => [], 'popups' => [], 'tips' => [], 'toasts' => [], 'scoreboard' => [],
		'bossBar' => [], 'sounds' => [], 'packets' => ['class'], 'clearInbox' => [],
		'position' => [], 'isOnline' => [], 'isAlive' => [], 'disconnectReason' => [], 'state' => [],
	];

	public static function execute(Agent $agent, Golem $golem, string $op, array $args) : mixed{
		if(!array_key_exists($op, self::METHODS)){
			throw new \InvalidArgumentException("Unknown player action $op");
		}
		foreach(array_keys($args) as $key){
			if(!in_array($key, self::METHODS[$op], true)){
				throw new \InvalidArgumentException("Unknown argument $key for $op");
			}
		}
		$player = $golem->player();
		return match($op){
			'chat' => $golem->chat(self::text($args, 'message')),
			'command' => $golem->command(self::text($args, 'command')),
			'grant' => $golem->grant(self::text($args, 'permission'), self::boolean($args, 'value', true)),
			'deny' => $golem->deny(self::text($args, 'permission')),
			'gamemode' => $golem->gamemode(GameMode::fromString(self::text($args, 'mode')) ?? throw new \InvalidArgumentException('Unknown gamemode')),
			'teleport' => $golem->teleport(self::vector($args)),
			'walkTo' => $golem->walkTo(self::vector($args), self::speed($args)),
			'walk' => $golem->walk(self::number($args, 'x'), self::number($args, 'z'), self::speed($args)),
			'sneak', 'sprint' => $golem->$op(self::boolean($args, 'value', true)),
			'look' => $player->setRotation(self::number($args, 'yaw'), self::number($args, 'pitch')),
			'give', 'hold' => $golem->$op(self::item($args)),
			'inventory' => ['main' => $player->getInventory(), 'armor' => $player->getArmorInventory(), 'offhand' => $player->getOffHandInventory(), 'cursor' => $player->getCursorInventory()],
			'setSlot' => self::setSlot($golem, $args),
			'breakBlock' => $golem->breakBlock(self::vector($args)),
			'interactBlock' => $golem->interactBlock(self::vector($args), self::integer($args, 'face', 1, 0, 5)),
			'attack', 'interactEntity' => $golem->$op(self::target($agent, $golem, $args)),
			'quit' => $golem->quit(isset($args['reason']) ? self::text($args, 'reason') : 'Golem agent quit'),
			'health' => isset($args['value']) ? $player->setHealth(self::number($args, 'value')) : $player->getHealth(),
			'hunger' => isset($args['value']) ? $player->getHungerManager()->setFood(self::number($args, 'value')) : $player->getHungerManager()->getFood(),
			'kill' => $player->kill(),
			'clickSlot' => $golem->clickSlot(self::integer($args, 'slot', null, 0, 255)),
			'waitForWindow' => $agent->clock()->until(fn() => $golem->window() !== null && $agent->getServer()->getTick() - $golem->session()->lastContainerOpenTick() >= 2 ? $golem->window() : null, self::integer($args, 'ticks', 60, 1, 12000), 'settled inventory window'),
			'submitForm' => $golem->submitForm(array_key_exists('data', $args) ? $args['data'] : throw new \InvalidArgumentException('data is required')),
			'clickButton' => $golem->clickButton($args['button'] ?? throw new \InvalidArgumentException('button is required')),
			'packets' => $golem->packets(isset($args['class']) ? self::text($args, 'class') : null),
			'state' => self::state($golem),
			default => $golem->$op(),
		};
	}

	public static function state(Golem $golem) : array{
		$p = $golem->player();
		return ['name' => $golem->name(), 'online' => $golem->isOnline(), 'alive' => $golem->isAlive(), 'health' => $p->getHealth(), 'hunger' => $p->getHungerManager()->getFood(), 'position' => View::value($golem->position()), 'world' => $p->getWorld()->getFolderName(), 'gamemode' => $p->getGamemode()->name, 'op' => $p->getServer()->isOp($golem->name()), 'sneaking' => $p->isSneaking(), 'sprinting' => $p->isSprinting(), 'messages' => $golem->messages(), 'form' => $golem->form() === null ? null : $golem->formData(), 'window' => View::value($golem->window()), 'inventory' => View::value($p->getInventory())];
	}

	public static function text(array $args, string $key) : string{
		$value = $args[$key] ?? null;
		if(!is_string($value) || $value === '' || strlen($value) > 32768){ throw new \InvalidArgumentException("$key must be a nonempty string of at most 32768 bytes"); }
		return $value;
	}

	public static function integer(array $args, string $key, ?int $default, int $min, int $max) : int{
		$value = $args[$key] ?? $default;
		if(!is_int($value) || $value < $min || $value > $max){ throw new \InvalidArgumentException("$key must be an integer between $min and $max"); }
		return $value;
	}

	private static function boolean(array $args, string $key, bool $default) : bool{
		$value = $args[$key] ?? $default;
		if(!is_bool($value)){ throw new \InvalidArgumentException("$key must be boolean"); }
		return $value;
	}

	private static function number(array $args, string $key) : float{
		$value = $args[$key] ?? null;
		if((!is_int($value) && !is_float($value)) || !is_finite((float) $value)){ throw new \InvalidArgumentException("$key must be finite numeric data"); }
		return (float) $value;
	}

	private static function speed(array $args) : float{
		$value = self::number($args + ['speed' => Golem::WALK_SPEED], 'speed');
		if($value <= 0 || $value > 100){ throw new \InvalidArgumentException('speed must be greater than zero and at most 100'); }
		return $value;
	}

	private static function vector(array $args) : Vector3{
		return new Vector3(self::number($args, 'x'), self::number($args, 'y'), self::number($args, 'z'));
	}

	private static function item(array $args) : Item{
		$item = StringToItemParser::getInstance()->parse(self::text($args, 'item')) ?? throw new \InvalidArgumentException('Unknown item');
		return $item->setCount(self::integer($args, 'count', 1, 1, $item->getMaxStackSize()));
	}

	private static function setSlot(Golem $golem, array $args) : null{
		$inventory = $golem->player()->getInventory();
		$inventory->setItem(self::integer($args, 'slot', null, 0, $inventory->getSize() - 1), self::item($args));
		return null;
	}

	private static function target(Agent $agent, Golem $golem, array $args) : Entity|Golem{
		$target = $args['target'] ?? null;
		if(is_string($target)){ return $agent->golem($target); }
		if(is_int($target)){ return $golem->player()->getWorld()->getEntity($target) ?? throw new \InvalidArgumentException('Entity not found in this world'); }
		throw new \InvalidArgumentException('target must be a golem name or entity runtime ID');
	}
}
