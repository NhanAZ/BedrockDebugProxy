<?php

declare(strict_types=1);

use BedrockDebugProxy\GolemAgent\Agent;
use pocketmine\block\VanillaBlocks;
use pocketmine\event\entity\EntityDamageEvent;
use pocketmine\event\entity\EntityDamageByEntityEvent;
use pocketmine\math\Vector3;

return static function(Agent $agent, array $params) : Generator{
	$checks = [];
	$check = static function(bool $condition, string $name) use (&$checks) : void{
		if(!$condition){ throw new RuntimeException('Behaviour assertion failed: ' . $name); }
		$checks[] = $name;
	};
	$call = static fn(string $op, array $args = [], string $name = 'Thinker') => $agent->dispatch(['op' => $op, 'player' => $name, 'args' => $args]);
	$call('reset');
	$one = yield $call('spawn', ['name' => 'Thinker', 'profile' => 'curious', 'seed' => 42, 'traits' => ['curiosity' => 1, 'sociability' => 0]]);
	$two = yield $call('spawn', ['name' => 'Companion', 'profile' => 'calm', 'seed' => 43]);
	try{
		$world = $one->player()->getWorld();
		for($x = 248; $x <= 264; ++$x){ for($z = 248; $z <= 264; ++$z){
			$world->setBlockAt($x, 3, $z, VanillaBlocks::STONE());
			for($y = 4; $y <= 7; ++$y){ $world->setBlockAt($x, $y, $z, VanillaBlocks::AIR()); }
		} }
		$call('teleport', ['x' => 256.5, 'y' => 4, 'z' => 256.5]);
		$call('teleport', ['x' => 258.5, 'y' => 4, 'z' => 256.5], 'Companion');
		$check($call('personality')['seed'] === 42, 'spawn retains seed');
		$check($call('state')['behaviour']['personality']['profile'] === 'curious', 'state includes personality');
		$check(!$call('autonomy')['enabled'], 'autonomy starts disabled');
		$call('autonomy', ['enabled' => true, 'ticks' => 100, 'radius' => 4]);
		yield $agent->clock()->wait(10);
		$state = $call('state');
		$check($state['behaviour']['decisions'] > 0, 'autonomy makes decisions');
		$check($state['behaviour']['enabled'], 'observation does not pause autonomy');
		$call('look', ['yaw' => 90, 'pitch' => 0]);
		$check(!$call('autonomy')['enabled'], 'manual control pauses autonomy');
		$move = $call('move', ['x' => 260, 'y' => 4, 'z' => 256.5, 'speed' => 0.5]);
		$check($move['movement']['status'] === 'pending', 'move returns before arrival');
		yield $agent->clock()->wait(2);
		$check($call('state')['behaviour']['movement']['status'] === 'pending', 'query moving player');
		$call('chat', ['message' => 'Walking while talking']);
		$call('sprint', ['value' => true]);
		$check($call('state')['behaviour']['movement']['status'] === 'pending', 'manual chat and sprint preserve manual movement');
		$call('jump');
		$check($call('state')['behaviour']['movement']['status'] === 'pending', 'manual jump preserves manual movement');
		$call('stop');
		$check($call('autonomy')['movement']['status'] === 'cancelled', 'stop cancels movement');
		$stopped = $one->position()->asVector3();
		yield $agent->clock()->wait(5);
		$check(abs($one->position()->x - $stopped->x) < 0.001 && abs($one->position()->z - $stopped->z) < 0.001, 'stop prevents continued horizontal movement');
		yield $agent->clock()->wait(20);
		$call('move', ['x' => 260, 'y' => 4, 'z' => 256.5, 'speed' => 0.5]);
		$call('move', ['x' => 255, 'y' => 4, 'z' => 256.5, 'speed' => 4]);
		yield $agent->clock()->wait(15);
		$check($call('autonomy')['movement']['status'] !== 'failed', 'replacement ignores stale movement callback');
		$call('stop');
		$call('personality', ['profile' => 'social', 'seed' => 8, 'traits' => ['sociability' => 1]]);
		$call('clearInbox');
		$call('autonomy', ['enabled' => true, 'ticks' => 10]);
		yield $agent->clock()->wait(12);
		$check(str_contains(implode(' ', $call('messages')), 'Hello, Companion.'), 'social personality greets nearby player');
		$check(!$call('autonomy')['enabled'], 'autonomy expires by tick deadline');
		$call('personality', ['profile' => 'calm', 'seed' => 10, 'traits' => ['curiosity' => 0, 'sociability' => 0]]);
		$call('autonomy', ['enabled' => true, 'ticks' => 100]);
		yield $agent->clock()->wait(2);
		$check($call('autonomy')['last_action'] === 'idle', 'calm personality idles');
		$one->player()->attack(new EntityDamageEvent($one->player(), EntityDamageEvent::CAUSE_CUSTOM, 1));
		yield $agent->clock()->wait(2);
		$check($call('autonomy')['last_action'] === 'retreat', 'hurt cautious player retreats');
		$call('stop');
		$call('personality', ['profile' => 'bold', 'seed' => 10, 'traits' => ['curiosity' => 0, 'sociability' => 0]]);
		$call('teleport', ['x' => 256.5, 'y' => 4, 'z' => 256.5]);
		$call('teleport', ['x' => 258, 'y' => 4, 'z' => 256.5], 'Companion');
		yield $agent->clock()->wait(20);
		$one->player()->noDamageTicks = $two->player()->noDamageTicks = 0;
		$call('autonomy', ['enabled' => true, 'ticks' => 100]);
		$otherHealth = $two->player()->getHealth();
		$one->player()->attack(new EntityDamageByEntityEvent($two->player(), $one->player(), EntityDamageEvent::CAUSE_ENTITY_ATTACK, 1));
		yield $agent->clock()->wait(2);
		$check($two->player()->getHealth() === $otherHealth, 'autonomy does not attack without configured target');
		$call('stop');
		$call('teleport', ['x' => 256.5, 'y' => 4, 'z' => 256.5]);
		yield $agent->clock()->wait(20);
		$call('autonomy', ['enabled' => true, 'ticks' => 100, 'combatTarget' => 'Companion']);
		$one->player()->attack(new EntityDamageByEntityEvent($two->player(), $one->player(), EntityDamageEvent::CAUSE_ENTITY_ATTACK, 1));
		yield $agent->clock()->wait(2);
		$check($two->player()->getHealth() < $otherHealth, 'bold personality retaliates against configured attacker');
		$call('stop');
		$call('teleport', ['x' => 256.5, 'y' => 4, 'z' => 256.5]);
		$call('personality', ['profile' => 'curious', 'seed' => 4, 'traits' => ['curiosity' => 1, 'sociability' => 0]]);
		for($x = 252; $x <= 260; ++$x){ for($z = 252; $z <= 260; ++$z){
			if($x !== 256 || $z !== 256){ $world->setBlockAt($x, 3, $z, VanillaBlocks::AIR()); }
		} }
		$call('autonomy', ['enabled' => true, 'ticks' => 10]);
		yield $agent->clock()->wait(3);
		$check($call('autonomy')['movement']['status'] !== 'pending', 'autonomy avoids unsupported ground');
		$check(abs($one->position()->x - 256.5) < 0.01 && abs($one->position()->z - 256.5) < 0.01, 'terrain avoidance preserves location');
		$call('stop');
		for($x = 252; $x <= 260; ++$x){ for($z = 252; $z <= 260; ++$z){ $world->setBlockAt($x, 3, $z, VanillaBlocks::STONE()); } }
		$call('autonomy', ['enabled' => true, 'ticks' => 100]);
		$call('kill');
		$check(!$call('autonomy')['enabled'], 'manual death stops autonomy');
		$call('respawn');
		foreach([['personality', ['traits' => ['pace' => 5]]], ['autonomy', ['enabled' => true, 'radius' => 100]], ['autonomy', ['enabled' => true, 'combatTarget' => 'Thinker']]] as [$op, $args]){
			try{ $call($op, $args); throw new RuntimeException('Invalid behaviour input accepted'); }
			catch(InvalidArgumentException){ $check(true, 'reject invalid ' . $op); }
		}
		return ['passed' => count($checks), 'checks' => $checks];
	}finally{ $call('reset'); }
};
