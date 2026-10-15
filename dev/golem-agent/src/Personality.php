<?php

declare(strict_types=1);

namespace BedrockDebugProxy\GolemAgent;

use Random\Engine\Xoshiro256StarStar;
use Random\Randomizer;

// Seeded choices are repeatable for the same sequence of observations.
final class Personality{
	public const PROFILES = [
		'balanced' => ['pace' => 0.5, 'curiosity' => 0.5, 'sociability' => 0.5, 'boldness' => 0.5],
		'calm' => ['pace' => 0.2, 'curiosity' => 0.25, 'sociability' => 0.4, 'boldness' => 0.15],
		'curious' => ['pace' => 0.6, 'curiosity' => 0.9, 'sociability' => 0.5, 'boldness' => 0.45],
		'bold' => ['pace' => 0.85, 'curiosity' => 0.65, 'sociability' => 0.3, 'boldness' => 0.9],
		'social' => ['pace' => 0.4, 'curiosity' => 0.4, 'sociability' => 0.95, 'boldness' => 0.3],
	];
	private Randomizer $random;
	public readonly array $traits;

	public function __construct(public readonly string $profile, public readonly int $seed, array $overrides = []){
		$base = self::PROFILES[$profile] ?? throw new \InvalidArgumentException('Unknown personality profile');
		$this->random = new Randomizer(new Xoshiro256StarStar($seed));
		foreach($overrides as $key => $value){
			if(!array_key_exists($key, $base) || (!is_int($value) && !is_float($value)) || !is_finite((float) $value) || $value < 0 || $value > 1){
				throw new \InvalidArgumentException('Personality traits must be known names with numeric values from 0 to 1');
			}
		}
		foreach($base as $key => $value){ $base[$key] = $overrides[$key] ?? max(0.0, min(1.0, $value + $this->integer(-10, 10) / 100)); }
		$this->traits = $base;
	}

	public function integer(int $min, int $max) : int{ return $this->random->getInt($min, $max); }
	public function chance(float $probability) : bool{ return $this->integer(0, 9999) < $probability * 10000; }
	public function speed() : float{ return 2.0 + 4.0 * $this->traits['pace']; }
	public function delay() : int{ return 20 + (int) round(60 * (1 - $this->traits['pace'])) + $this->integer(0, 20); }
	public function snapshot() : array{ return ['profile' => $this->profile, 'seed' => $this->seed, 'traits' => $this->traits]; }

	public function choose(bool $hurt, bool $lowHealth, bool $blocked, bool $nearby) : string{
		if($hurt){ return !$lowHealth && $this->traits['boldness'] >= 0.6 ? 'defend' : 'retreat'; }
		if($blocked){ return $this->chance($this->traits['curiosity']) ? 'jump' : 'look'; }
		if($nearby && $this->chance($this->traits['sociability'])){ return 'greet'; }
		return $this->chance($this->traits['curiosity']) ? 'wander' : 'idle';
	}
}
