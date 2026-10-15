<?php

declare(strict_types=1);

require dirname(__DIR__) . '/src/Personality.php';

use BedrockDebugProxy\GolemAgent\Personality;

$checks = 0;
$check = static function(bool $value) use (&$checks) : void{
	if(!$value){ throw new RuntimeException('Personality regression failed at check ' . $checks); }
	++$checks;
};
$a = new Personality('curious', 123);
$b = new Personality('curious', 123);
$check($a->snapshot() === $b->snapshot());
for($i = 0; $i < 100; ++$i){
	$check($a->choose(false, false, $i % 3 === 0, $i % 2 === 0) === $b->choose(false, false, $i % 3 === 0, $i % 2 === 0));
	$check($a->delay() === $b->delay());
}
$check((new Personality('calm', 1))->speed() < (new Personality('bold', 1))->speed());
$check((new Personality('calm', 1))->choose(true, false, false, false) === 'retreat');
$check((new Personality('bold', 1))->choose(true, false, false, false) === 'defend');
$check((new Personality('bold', 1))->choose(true, true, false, false) === 'retreat');
$check((new Personality('balanced', 1))->snapshot() !== (new Personality('balanced', 2))->snapshot());
$check((new Personality('balanced', 1, ['curiosity' => 0]))->choose(false, false, false, false) === 'idle');
$check((new Personality('balanced', 1, ['curiosity' => 1]))->choose(false, false, false, false) === 'wander');
$check((new Personality('social', 1, ['sociability' => 1]))->choose(false, false, false, true) === 'greet');
foreach([['unknown', []], ['calm', ['pace' => 2]], ['calm', ['pace' => '0.5']], ['calm', ['missing' => 0.5]], ['calm', ['pace' => NAN]]] as [$profile, $traits]){
	try{ new Personality($profile, 1, $traits); throw new RuntimeException('Invalid personality accepted'); }
	catch(InvalidArgumentException){ ++$checks; }
}
echo json_encode(['passed' => $checks], JSON_THROW_ON_ERROR) . "\n";
