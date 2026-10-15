<?php

declare(strict_types=1);

use BedrockDebugProxy\GolemAgent\Agent;
use pocketmine\block\VanillaBlocks;
use pocketmine\event\Event;
use pocketmine\event\EventPriority;
use pocketmine\event\Cancellable;
use pocketmine\event\CancellableTrait;
use pocketmine\event\HandlerListManager;
use pocketmine\event\block\BlockBreakEvent;
use pocketmine\event\inventory\InventoryTransactionEvent;
use pocketmine\event\player\PlayerChatEvent;
use pocketmine\event\player\PlayerJoinEvent;
use pocketmine\event\player\PlayerMoveEvent;
use pocketmine\event\player\PlayerQuitEvent;
use pocketmine\form\Form;
use pocketmine\block\inventory\ChestInventory;
use pocketmine\item\VanillaItems;
use pocketmine\math\Vector3;
use pocketmine\player\Player;
use pocketmine\world\Position;
use pocketmine\network\mcpe\protocol\BossEventPacket;
use pocketmine\network\mcpe\protocol\SetDisplayObjectivePacket;
use pocketmine\network\mcpe\protocol\PlaySoundPacket;

if(!class_exists('BDPGolemFixtureEvent', false)){
	final class BDPGolemFixtureEvent extends Event implements Cancellable{
		use CancellableTrait;
	}
}

// Trusted integration fixture for an isolated flat development world.
return static function(Agent $agent, array $params) : Generator{
	$checks = [];
	$check = static function(bool $condition, string $name) use (&$checks) : void{
		if(!$condition){ throw new RuntimeException("Integration assertion failed: $name"); }
		$checks[] = $name;
	};
	$call = static fn(string $op, array $args = [], string $player = 'AgentOne') => $agent->dispatch(['op' => $op, 'player' => $player, 'args' => $args]);
	$call('reset');
	foreach([PlayerJoinEvent::class, PlayerChatEvent::class, PlayerMoveEvent::class, PlayerQuitEvent::class, BlockBreakEvent::class, InventoryTransactionEvent::class] as $class){
		$call('subscribe', ['class' => $class]);
	}
	$start = $call('events')['next'];
	$call('subscribe', ['class' => BDPGolemFixtureEvent::class]);
	$custom = new BDPGolemFixtureEvent(); $custom->cancel(); $custom->call();
	$customEvents = $call('events', ['after' => $start])['events'];
	$check(count($customEvents) === 1 && $customEvents[0]['class'] === BDPGolemFixtureEvent::class && $customEvents[0]['cancelled'] === true, 'custom cancelled event observation');
	$one = yield $call('spawn', ['name' => 'AgentOne']);
	$two = yield $call('spawn', ['name' => 'AgentTwo']);
	try{
		$check($one->isOnline() && $two->isOnline(), 'two players joined');
		$check(count($call('list')) === 2, 'list players');
		$call('chat', ['message' => 'golem integration chat']);
		$check(str_contains(implode('\n', $call('messages')), 'golem integration chat'), 'chat event and inbox');
		$check($call('command', ['command' => 'help']) === true, 'player command dispatch');
		$call('op'); $check($one->player()->getServer()->isOp($one->name()), 'grant temporary operator');
		$call('deop'); $check(!$one->player()->getServer()->isOp($one->name()), 'revoke operator');
		$call('grant', ['permission' => 'golem.test.permission']); $check($one->player()->hasPermission('golem.test.permission'), 'grant permission');
		$call('deny', ['permission' => 'golem.test.permission']); $check(!$one->player()->hasPermission('golem.test.permission'), 'deny permission');
		$call('gamemode', ['mode' => 'creative']); $check($call('state')['gamemode'] === 'CREATIVE', 'gamemode');
		$call('gamemode', ['mode' => 'survival']);
		$p = $one->player(); $world = $p->getWorld();
		// Build a deterministic isolated ground area using the server API.
		for($x = 248; $x <= 264; ++$x){ for($z = 248; $z <= 264; ++$z){
			$world->setBlockAt($x, 3, $z, VanillaBlocks::STONE());
			for($y = 4; $y <= 7; ++$y){ $world->setBlockAt($x, $y, $z, VanillaBlocks::AIR()); }
		} }
		$call('teleport', ['x' => 256.5, 'y' => 4, 'z' => 256.5]);
		$call('teleport', ['x' => 257.5, 'y' => 4, 'z' => 256.5], 'AgentTwo');
		yield $call('wait', ['ticks' => 5]);
		yield $call('walk', ['x' => 1, 'z' => 0]);
		$check(abs($one->position()->x - 257.5) < 0.3, 'walk with movement events');
		yield $call('walkTo', ['x' => 256.5, 'y' => 4, 'z' => 256.5]);
		$check(abs($one->position()->x - 256.5) < 0.3, 'walkTo');
		$check($call('jump') === true, 'jump'); yield $call('wait', ['ticks' => 20]);
		$call('sneak'); $check($p->isSneaking(), 'sneak'); $call('sneak', ['value' => false]);
		$call('sprint'); $check($p->isSprinting(), 'sprint'); $call('sprint', ['value' => false]);
		$call('look', ['yaw' => 90, 'pitch' => 0]); $check($p->getLocation()->yaw === 90.0, 'look');
		$call('give', ['item' => 'bread', 'count' => 3]); $check($p->getInventory()->contains(VanillaItems::BREAD()->setCount(3)), 'give item');
		$call('setSlot', ['slot' => 5, 'item' => 'apple', 'count' => 2]); $check($p->getInventory()->getItem(5)->getCount() === 2, 'set inventory slot');
		$call('hold', ['item' => 'diamond_pickaxe']);
		$world->setBlockAt(256, 4, 257, VanillaBlocks::STONE());
		$call('op');
		$protection = $agent->getServer()->getPluginManager()->registerEvent(BlockBreakEvent::class, static function(BlockBreakEvent $event) : void{ $event->cancel(); }, EventPriority::NORMAL, $agent, true);
		try{
			$check($call('breakBlock', ['x' => 256, 'y' => 4, 'z' => 257]) === false && !$world->getBlockAt(256, 4, 257)->isTransparent(), 'plugin cancellation prevents block break');
		}finally{ HandlerListManager::global()->unregisterAll($protection); }
		$check($call('breakBlock', ['x' => 256, 'y' => 4, 'z' => 257]) === true, 'break block through player');
		$call('hold', ['item' => 'stone', 'count' => 3]);
		$check($call('interactBlock', ['x' => 256, 'y' => 3, 'z' => 257, 'face' => 1]) === true, 'place held block');
		$call('hold', ['item' => 'apple']); $call('hunger', ['value' => 10]);
		$check(is_bool($call('useItem')), 'use held item');
		$call('health', ['value' => 12]); $check($call('health') === 12.0, 'set health');
		$check($call('hunger') <= 10, 'set hunger');
		$call('health', ['value' => 20]);
		$call('hold', ['item' => 'diamond_sword']);
		yield $call('wait', ['ticks' => 65]);
		$victimHealth = $two->player()->getHealth();
		$check($call('attack', ['target' => 'AgentTwo']) === true && $two->player()->getHealth() < $victimHealth, 'attack damages player');
		$check(is_bool($call('interactEntity', ['target' => 'AgentTwo'])), 'interact entity');
		$form = new class implements Form{
			public mixed $answer = null;
			public function jsonSerialize() : array{ return ['type' => 'form', 'title' => 'Agent menu', 'content' => '', 'buttons' => [['text' => 'Confirm']]]; }
			public function handleResponse(Player $player, $data) : void{ $this->answer = $data; $player->sendMessage('form confirmed'); }
		};
		$p->sendForm($form);
		$check($call('formData')['title'] === 'Agent menu', 'inspect form');
		$call('clickButton', ['button' => 'Confirm']); $check($form->answer === 0, 'form button by label');
		$p->sendForm($form); $call('submitForm', ['data' => 0]); $check($form->answer === 0, 'submit form');
		$p->sendForm($form); $call('closeForm'); $check($form->answer === null, 'close form');
		$window = new ChestInventory(new Position(256, 4, 258, $world)); $window->setItem(0, VanillaItems::DIAMOND());
		$p->setCurrentWindow($window);
		yield $call('waitForWindow');
		$check($call('clickSlot', ['slot' => 0]) === true, 'inventory transaction');
		$check(!$p->getCursorInventory()->getItem(0)->isNull(), 'cursor acquired item');
		$call('closeWindow'); $check($p->getCurrentWindow() === null, 'close window');
		$p->sendTitle('Agent title', 'Agent subtitle'); $p->sendActionBarMessage('Agent action');
		$p->sendTip('Agent tip'); $p->sendPopup('Agent popup'); $p->sendToastNotification('Agent toast', 'body');
		foreach(['titles', 'subtitles', 'actionBars', 'tips', 'popups', 'toasts'] as $op){ $check(count($call($op)) > 0, $op); }
		$p->getNetworkSession()->sendDataPacket(BossEventPacket::show(9876, 'Agent boss', 0.5));
		$check($call('bossBar')['title'] === 'Agent boss' && $call('bossBar')['progress'] === 0.5, 'boss bar content');
		$p->getNetworkSession()->sendDataPacket(SetDisplayObjectivePacket::create('sidebar', 'agent', 'Agent score', 'dummy', 0));
		$check($call('scoreboard')['title'] === 'Agent score', 'scoreboard content');
		$sound = new PlaySoundPacket();
		$sound->soundName = 'random.levelup'; $sound->x = 256; $sound->y = 4; $sound->z = 256;
		$sound->volume = 1; $sound->pitch = 1;
		// Defaults differ between supported server libraries. Set newer optional fields when exposed.
		foreach(['loopCount' => 1, 'bypassListenerRangeCheck' => false, 'serverSoundHandle' => null, 'playbackPositionSeconds' => null] as $key => $value){ if(property_exists($sound, $key)){ $sound->$key = $value; } }
		$p->getNetworkSession()->sendDataPacket($sound);
		$check(in_array('random.levelup', $call('sounds'), true), 'sound name observation');
		foreach(['scoreboard', 'bossBar', 'sounds', 'rawMessages', 'packets', 'lastMessage', 'position', 'inventory', 'window', 'disconnectReason'] as $op){ $call($op); $check(true, $op . ' observation'); }
		$call('assert', ['path' => 'online', 'equals' => true]); $check(true, 'state assertion');
		foreach([
			['assert', ['path' => 'online', 'equals' => false]],
			['assert', ['path' => 'online']],
			['walk', ['x' => 1, 'z' => 0, 'speed' => -1]],
			['give', ['item' => 'not_a_real_item']],
			['setSlot', ['slot' => 1000, 'item' => 'stone']],
			['sneak', ['value' => 'false']],
			['reset', ['typo' => true]],
			['script', ['file' => '../outside.php']],
			['spawn', ['name' => 'AgentOne']],
		] as [$op, $args]){
			$rejected = false;
			try{ $call($op, $args); }catch(Throwable){ $rejected = true; }
			$check($rejected, 'reject invalid ' . $op);
		}
		yield $call('waitFor', ['path' => 'messages', 'contains' => 'form confirmed', 'ticks' => 10]); $check(true, 'wait for state');
		$call('kill'); $check(!$call('isAlive'), 'death');
		$check($call('respawn') === true, 'respawn'); yield $call('wait', ['ticks' => 5]);
		$check($call('isAlive'), 'alive after respawn');
		$call('quit', ['reason' => 'fixture quit'], 'AgentTwo'); $check(!$call('isOnline', [], 'AgentTwo'), 'quit');
		$events = $call('events', ['after' => $start])['events'];
		foreach([PlayerJoinEvent::class, PlayerChatEvent::class, PlayerMoveEvent::class, PlayerQuitEvent::class, BlockBreakEvent::class, InventoryTransactionEvent::class] as $class){
			$check(in_array($class, array_column($events, 'class'), true), 'observed ' . $class);
		}
		$call('clearInbox'); $check($call('messages') === [] && $call('packets') === [], 'explicit inbox clear');
		return ['passed' => count($checks), 'checks' => $checks, 'runtime' => Agent::RUNTIME, 'evidence' => 'server_side_simulation'];
	}finally{
		$call('reset');
	}
};
