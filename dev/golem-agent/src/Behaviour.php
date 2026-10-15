<?php

declare(strict_types=1);

namespace BedrockDebugProxy\GolemAgent;

use Golem\Golem;
use Golem\Runtime\Coroutine\Deferred;
use pocketmine\block\Liquid;
use pocketmine\math\Vector3;

// One bounded controller per owned player. Direct commands always take control.
final class Behaviour{
	private bool $enabled = false;
	private int $until = 0;
	private int $next = 0;
	private int $lastGreeting = -200;
	private float $radius = 8;
	private ?string $combatTarget = null;
	private Vector3 $home;
	private string $world;
	private ?Deferred $motion = null;
	private int $motionGeneration = 0;
	private string $motionStatus = 'idle';
	private ?string $motionError = null;
	private ?Vector3 $destination = null;
	private ?string $attacker = null;
	private bool $hurt = false;
	private bool $blocked = false;
	private string $last = 'idle';
	private int $decisions = 0;

	public function __construct(private Agent $agent, private Golem $golem, private Personality $personality){
		$this->home = $golem->position()->asVector3();
		$this->world = $golem->player()->getWorld()->getFolderName();
	}

	public function personality(array $args) : array{
		if($args !== []){
			$profile = isset($args['profile']) ? Actions::text($args, 'profile') : $this->personality->profile;
			$seed = Actions::integer($args, 'seed', $this->personality->seed, 0, 2147483647);
			$traits = $args['traits'] ?? [];
			if(!is_array($traits)){ throw new \InvalidArgumentException('traits must be an object'); }
			$replacement = new Personality($profile, $seed, $traits);
			$this->stop('personality changed');
			$this->personality = $replacement;
		}
		return $this->personality->snapshot();
	}

	public function configure(array $args) : array{
		if($args === []){ return $this->snapshot(); }
		if(!isset($args['enabled']) || !is_bool($args['enabled'])){ throw new \InvalidArgumentException('enabled must be boolean'); }
		$ticks = Actions::integer($args, 'ticks', 200, 1, 12000);
		$radius = Actions::integer($args, 'radius', 8, 1, 32);
		$target = isset($args['combatTarget']) ? Actions::text($args, 'combatTarget') : null;
		if($target !== null){
			$other = $this->agent->golem($target);
			if($other === $this->golem || !$other->isOnline()){ throw new \InvalidArgumentException('combatTarget must be another online owned golem'); }
		}
		$this->stop($args['enabled'] ? 'autonomy restarted' : 'autonomy disabled');
		if($args['enabled']){
			if(!$this->golem->isOnline() || !$this->golem->isAlive()){ throw new \RuntimeException('Autonomy requires an online living golem'); }
			$this->enabled = true;
			$this->until = $this->agent->getServer()->getTick() + $ticks;
			$this->next = $this->agent->getServer()->getTick() + 1;
			$this->home = $this->golem->position()->asVector3();
			$this->world = $this->golem->player()->getWorld()->getFolderName();
			$this->radius = $radius;
			$this->combatTarget = $target === null ? null : strtolower($target);
			$this->hurt = $this->blocked = false;
			$this->attacker = null;
		}
		return $this->snapshot();
	}

	public function snapshot() : array{
		return ['enabled' => $this->enabled, 'until_tick' => $this->until, 'radius' => $this->radius, 'home' => View::value($this->home), 'combat_target' => $this->combatTarget, 'last_action' => $this->last, 'decisions' => $this->decisions, 'movement' => ['status' => $this->motionStatus, 'error' => $this->motionError, 'target' => View::value($this->destination)], 'personality' => $this->personality->snapshot()];
	}

	public function stop(string $reason = 'operator stop') : array{
		$changed = $this->enabled || $this->motion?->isPending();
		$this->enabled = false;
		++$this->motionGeneration;
		if($this->motion?->isPending()){
			$this->golem->movement()->abort(new \RuntimeException($reason));
			$this->motionStatus = 'cancelled';
			$this->motionError = $reason;
		}
		$this->motion = null;
		if($changed){
			$this->golem->sprint(false);
			$this->last = 'stopped';
			$this->agent->observeBehaviour($this->golem->name(), ['action' => 'stop', 'reason' => $reason]);
		}
		return $this->snapshot();
	}

	public function takeControl(string $op) : void{
		if($this->enabled || in_array($op, ['move', 'walk', 'walkTo', 'teleport', 'stop', 'quit', 'kill', 'respawn'], true)){
			$this->stop('direct action ' . $op);
		}
	}

	public function damaged(?string $attacker) : void{
		if($this->enabled){ $this->hurt = true; $this->attacker = $attacker; $this->next = 0; }
	}

	public function move(Vector3 $target, float $speed) : array{
		$generation = ++$this->motionGeneration;
		$this->destination = clone $target;
		$this->motionStatus = 'pending'; $this->motionError = null;
		$this->motion = $this->golem->walkTo($target, $speed);
		$this->motion->then(function() use ($generation) : void{
			if($generation === $this->motionGeneration){ $this->motionStatus = 'arrived'; }
		}, function(\Throwable $error) use ($generation) : void{
			if($generation === $this->motionGeneration){
				$this->motionStatus = 'failed'; $this->motionError = $error->getMessage(); $this->blocked = true;
				$this->agent->observeBehaviour($this->golem->name(), ['action' => 'movement.failed', 'error' => $error->getMessage()]);
			}
		});
		return $this->snapshot();
	}

	public function tick() : void{
		if(!$this->enabled){ return; }
		$now = $this->agent->getServer()->getTick();
		$player = $this->golem->player();
		if(!$this->golem->isOnline() || !$this->golem->isAlive() || $now >= $this->until || $player->getWorld()->getFolderName() !== $this->world){ $this->stop('autonomy ended'); return; }
		if($now < $this->next || ($this->motion?->isPending() && !$this->hurt)){ return; }
		try{
			if($this->hurt && $this->motion?->isPending()){
				++$this->motionGeneration;
				$this->golem->movement()->abort(new \RuntimeException('Reacting to damage'));
				$this->motion = null; $this->motionStatus = 'cancelled'; $this->motionError = 'Reacting to damage';
			}
			$nearby = $this->agent->nearestGolem($this->golem, 6);
			$action = $this->personality->choose($this->hurt, $player->getHealth() <= 6, $this->blocked, $nearby !== null && $now - $this->lastGreeting >= 200);
			$this->next = $now + $this->personality->delay();
			$this->last = $action; ++$this->decisions;
			$this->agent->observeBehaviour($this->golem->name(), ['action' => $action, 'hurt' => $this->hurt, 'attacker' => $this->attacker, 'blocked' => $this->blocked, 'nearby' => $nearby?->name(), 'personality' => $this->personality->snapshot()]);
			$this->hurt = $this->blocked = false;
			switch($action){
				case 'defend':
					if($this->combatTarget !== null && $this->attacker === $this->combatTarget){
						$target = $this->agent->golem($this->combatTarget);
						if($target->isOnline() && $target->isAlive() && $target->player()->getWorld() === $player->getWorld() && $target->position()->distanceSquared($this->golem->position()) <= 9){
							$player->lookAt($target->position());
							$this->agent->observeBehaviour($this->golem->name(), ['action' => 'attack', 'target' => $target->name(), 'accepted' => $this->golem->attack($target)]);
							break;
						}
					}
					// No authorized in-range target. Fall back to avoidance.
				case 'retreat': $this->wander(true, $nearby); break;
				case 'wander': $this->wander(false, $nearby); break;
				case 'jump': $this->golem->jump(); break;
				case 'greet':
					$this->lastGreeting = $now;
					$player->lookAt($nearby->position());
					$this->golem->chat('Hello, ' . $nearby->name() . '.');
					break;
				case 'look': $player->setRotation((float) $this->personality->integer(0, 359), 0); break;
				case 'idle': $this->golem->sprint(false); break;
			}
		}catch(\Throwable $error){
			$this->stop('behaviour error');
			$this->agent->observeBehaviour($this->golem->name(), ['action' => 'error', 'error' => $error->getMessage()]);
		}
	}

	private function wander(bool $retreat, ?Golem $nearby) : void{
		$from = $this->golem->position();
		$angle = deg2rad($this->personality->integer(0, 359));
		if($retreat && $nearby !== null){ $angle = atan2($from->z - $nearby->position()->z, $from->x - $nearby->position()->x); }
		$distance = $this->personality->integer(1, 3);
		$target = $from->add(cos($angle) * $distance, 0, sin($angle) * $distance);
		if($target->distanceSquared($this->home) > $this->radius ** 2){ $target = $this->home; }
		$delta = $target->subtractVector($from);
		$steps = max(1, (int) ceil($delta->length() * 2));
		$world = $this->golem->player()->getWorld();
		for($i = 1; $i <= $steps; ++$i){
			$point = $from->addVector($delta->multiply($i / $steps));
			$x = (int) floor($point->x); $y = (int) floor($point->y); $z = (int) floor($point->z);
			if(!$world->isChunkLoaded($x >> 4, $z >> 4) || !$world->getBlockAt($x, $y - 1, $z)->isSolid() || $world->getBlockAt($x, $y, $z)->isSolid() || $world->getBlockAt($x, $y + 1, $z)->isSolid() || $world->getBlockAt($x, $y, $z) instanceof Liquid){
				$this->blocked = true;
				$this->agent->observeBehaviour($this->golem->name(), ['action' => 'avoid', 'position' => View::value($point)]);
				return;
			}
		}
		$this->golem->sprint($retreat || $this->personality->traits['pace'] >= 0.7);
		$this->move($target, $retreat ? 6.0 : $this->personality->speed());
	}
}
