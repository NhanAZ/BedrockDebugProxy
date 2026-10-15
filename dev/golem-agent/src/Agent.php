<?php

declare(strict_types=1);

namespace BedrockDebugProxy\GolemAgent;

use Golem\Golem;
use Golem\Runtime\Clock;
use Golem\Runtime\Coroutine\Coroutine;
use Golem\Runtime\Coroutine\Deferred;
use Golem\Runtime\GolemFactory;
use Golem\Runtime\Network\GolemSession;
use pocketmine\event\Cancellable;
use pocketmine\event\Event;
use pocketmine\event\EventPriority;
use pocketmine\event\entity\EntityDamageEvent;
use pocketmine\event\entity\EntityDamageByEntityEvent;
use pocketmine\player\Player;
use pocketmine\event\player\PlayerDataSaveEvent;
use pocketmine\plugin\PluginBase;
use pocketmine\scheduler\ClosureTask;

final class Agent extends PluginBase{
	public const SCHEMA = 'bedrockdebugproxy.golem.v1';
	public const RUNTIME = 'achedon12/golem@18c86ea25b1d67fa9ceb6079c46ade2f8f99fc31+bdp-resource-pack-status-1';
	private const SESSION_ACTIONS = ['capabilities' => [], 'spawn' => ['name', 'profile', 'seed', 'traits'], 'list' => [], 'reset' => [], 'wait' => ['ticks'], 'subscribe' => ['class'], 'events' => ['after'], 'assert' => ['path', 'equals', 'contains'], 'waitFor' => ['path', 'equals', 'contains', 'ticks'], 'script' => ['file', 'params']];
	private const MAX_BYTES = 1048576;
	private const MAX_JOURNAL = 268435456;
	private GolemFactory $factory;
	private Clock $clock;
	/** @var array<string, Golem> */
	private array $golems = [];
	/** @var array<string, Behaviour> */
	private array $behaviours = [];
	private array $subscriptions = [];
	private array $events = [];
	private int $eventSequence = 0;
	private int $sequence = 0;
	private int $started;
	private int $journalBytes = 0;
	private string $boot;
	private string $journal;
	private ?array $active = null;
	private ?Coroutine $coroutine = null;
	private int $generation = 0;

	protected function onEnable() : void{
		$runtime = $this->getServer()->getPluginManager()->getPlugin('Golem');
		if($runtime?->getDescription()->getVersion() !== '0.7.0'){
			throw new \RuntimeException('GolemAgent requires the reviewed Golem 0.7.0 runtime. See docs/golem.md.');
		}
		$runtimeFile = (new \ReflectionClass(Golem::class))->getFileName();
		$runtimeRoot = substr($runtimeFile, 0, -strlen('src/Golem.php'));
		$lock = json_decode(file_get_contents(dirname(__DIR__) . '/runtime-lock.json'), true, 32, JSON_THROW_ON_ERROR);
		foreach($lock['sha256'] as $path => $expected){
			if(!is_file($runtimeRoot . $path) || hash_file('sha256', $runtimeRoot . $path) !== $expected){
				throw new \RuntimeException("Golem runtime source mismatch at $path. Build the pinned runtime with the documented patch.");
			}
		}
		$this->boot = bin2hex(random_bytes(16));
		$this->started = hrtime(true);
		$this->journal = $this->getDataFolder() . 'journal-' . $this->boot . '.jsonl';
		foreach(['requests', 'responses', 'scripts'] as $folder){
			$path = $this->getDataFolder() . $folder;
			if(!is_dir($path) && !mkdir($path, 0700, true)){ throw new \RuntimeException("Cannot create $path"); }
		}
		$this->factory = new GolemFactory($this);
		$this->clock = new Clock($this);
		$this->getServer()->getPluginManager()->registerEvent(PlayerDataSaveEvent::class, static function(PlayerDataSaveEvent $event) : void{
			if($event->getPlayer()?->getNetworkSession() instanceof GolemSession){ $event->cancel(); }
		}, EventPriority::MONITOR, $this, true);
		$this->getServer()->getPluginManager()->registerEvent(EntityDamageEvent::class, function(EntityDamageEvent $event) : void{
			$victim = $event->getEntity();
			if($event->isCancelled() || !$victim instanceof Player || $event->getFinalDamage() <= 0){ return; }
			$controller = $this->behaviours[strtolower($victim->getName())] ?? null;
			$damager = $event instanceof EntityDamageByEntityEvent ? $event->getDamager() : null;
			$controller?->damaged($damager instanceof Player ? strtolower($damager->getName()) : null);
		}, EventPriority::MONITOR, $this, false);
		$this->record('runtime.start', ['runtime' => self::RUNTIME, 'server' => $this->getServer()->getVersion(), 'limits' => ['message_bytes' => self::MAX_BYTES, 'journal_bytes' => self::MAX_JOURNAL, 'players' => 32, 'events_in_memory' => 4096, 'packets_per_player_inbox' => 16384], 'evidence' => 'server_side_simulation']);
		$this->write($this->getDataFolder() . 'ready.json', ['schema' => self::SCHEMA, 'boot' => $this->boot, 'runtime' => self::RUNTIME]);
		$this->getScheduler()->scheduleRepeatingTask(new ClosureTask(function() : void{
			try{ $this->pump(); }catch(\Throwable $error){
				$this->getLogger()->logException($error);
				$this->getServer()->getPluginManager()->disablePlugin($this);
			}
		}), 1);
	}

	protected function onDisable() : void{
		$ready = $this->getDataFolder() . 'ready.json';
		if(is_file($ready)){ unlink($ready); }
		$this->coroutine?->cancel();
		if(isset($this->factory)){ $this->resetGolems(); }
		if($this->active !== null){
			try{ $this->finish(null, new \RuntimeException('GolemAgent disabled')); }
			catch(\Throwable $error){ $this->getLogger()->logException($error); }
		}
	}

	public function clock() : Clock{ return $this->clock; }
	public function golem(string $name) : Golem{ return $this->golems[strtolower($name)] ?? throw new \InvalidArgumentException("Unknown golem $name"); }
	public function behaviour(Golem $golem) : Behaviour{ return $this->behaviours[strtolower($golem->name())]; }
	public function state(Golem $golem) : array{ return Actions::state($golem) + ['behaviour' => $this->behaviour($golem)->snapshot()]; }
	public function observeBehaviour(string $name, array $data) : void{ $this->record('behaviour', ['player' => $name] + $data); }

	public function nearestGolem(Golem $self, float $radius) : ?Golem{
		$nearest = null; $distance = $radius ** 2;
		foreach($this->golems as $other){
			if($other === $self || !$other->isOnline() || !$other->isAlive() || $other->player()->getWorld() !== $self->player()->getWorld()){ continue; }
			$current = $self->position()->distanceSquared($other->position());
			if($current < $distance){ $distance = $current; $nearest = $other; }
		}
		return $nearest;
	}

	private function resetGolems() : void{
		$this->behaviours = [];
		$this->factory->despawnAll();
		$this->golems = [];
	}

	private function pump() : void{
		foreach($this->golems as $golem){
			if(count($golem->packets()) > 16384){
				$this->record('runtime.limit', ['player' => $golem->name(), 'limit' => 'packets_per_player_inbox', 'value' => 16384]);
				throw new \LengthException('Golem packet inbox limit reached. Read packets and explicitly clearInbox between test phases.');
			}
			$this->behaviour($golem)->tick();
		}
		if($this->active !== null){
			$id = $this->active['transport_id'];
			if($this->expired($this->active) || is_file($this->getDataFolder() . "requests/$id.cancel")){
				$this->coroutine?->cancel();
				++$this->generation;
				$this->resetGolems();
				$this->finish(null, new \RuntimeException('Request interrupted. Simulated players were reset. Earlier effects are not rolled back.'));
			}
			return;
		}
		$files = glob($this->getDataFolder() . 'requests/*.json') ?: [];
		if($files === []){ return; }
		sort($files, SORT_STRING);
		$file = $files[0];
		$id = basename($file, '.json');
		if(!preg_match('/^[0-9a-f-]{36}$/D', $id)){ throw new \InvalidArgumentException('Invalid mailbox request filename'); }
		// Claim before execution so a crash cannot cause an automatic replay.
		if(!rename($file, $file . '.claimed')){ throw new \RuntimeException('Cannot claim request'); }
		if(filesize($file . '.claimed') > self::MAX_BYTES){ throw new \LengthException('Request exceeds 1 MiB'); }
		$envelope = json_decode(file_get_contents($file . '.claimed'), true, 32, JSON_THROW_ON_ERROR);
		if(!is_array($envelope) || ($envelope['transport_id'] ?? null) !== $id){ throw new \InvalidArgumentException('Request transport identity mismatch'); }
		$this->active = $envelope;
		$this->record('request', $envelope);
		try{
			if(($envelope['schema'] ?? null) !== self::SCHEMA || ($envelope['boot'] ?? null) !== $this->boot){ throw new \RuntimeException('Request belongs to another runtime boot or schema'); }
			if($this->expired($envelope) || is_file($this->getDataFolder() . "requests/$id.cancel")){ throw new \RuntimeException('Request expired or cancelled before execution'); }
			$result = $this->dispatch($envelope['request']);
			if($result instanceof \Generator){
				$this->coroutine = Coroutine::run($result);
				$result = $this->coroutine->completion();
			}
			if($result instanceof Deferred){
				$result->then(function(mixed $value) use ($id) : void{
					if(($this->active['transport_id'] ?? null) === $id){ $this->finish($value); }
				}, function(\Throwable $error) use ($id) : void{
					if(($this->active['transport_id'] ?? null) === $id){ $this->finish(null, $error); }
				});
			}else{ $this->finish($result); }
		}catch(\Throwable $error){ $this->finish(null, $error); }
	}

	private function expired(array $envelope) : bool{
		$deadline = $envelope['deadline_unix_ms'] ?? 0;
		return !is_int($deadline) || $deadline <= (int) floor(microtime(true) * 1000) || $deadline > (int) floor(microtime(true) * 1000) + 600000;
	}

	private function finish(mixed $result, ?\Throwable $error = null) : void{
		if($this->active === null){ return; }
		$envelope = $this->active;
		$response = ['schema' => self::SCHEMA, 'boot' => $this->boot, 'transport_id' => $envelope['transport_id'], 'id' => $envelope['request']['id'] ?? '', 'ok' => $error === null, 'tick' => $this->getServer()->getTick()];
		try{
			if($error !== null){ throw $error; }
			$response['result'] = View::value($result);
			if(strlen(json_encode($response, JSON_THROW_ON_ERROR)) > self::MAX_BYTES){ throw new \LengthException('Response exceeds 1 MiB. Use a focused script projection.'); }
		}catch(\Throwable $failure){
			unset($response['result']);
			$response['ok'] = false;
			$response['error'] = $failure::class . ': ' . $failure->getMessage();
		}
		$this->record('response', $response);
		$this->write($this->getDataFolder() . 'responses/' . $envelope['transport_id'] . '.json', $response);
		$this->active = null;
		$this->coroutine = null;
	}

	public function dispatch(array $request) : mixed{
		$op = $request['op'] ?? '';
		$args = $request['args'] ?? [];
		if(!is_array($args)){ throw new \InvalidArgumentException('args must be an object'); }
		if(isset(self::SESSION_ACTIONS[$op])){
			foreach(array_keys($args) as $key){
				if(!in_array($key, self::SESSION_ACTIONS[$op], true)){ throw new \InvalidArgumentException("Unknown argument $key for $op"); }
			}
		}
		switch($op){
			case 'capabilities':
				return ['runtime' => self::RUNTIME, 'evidence' => 'server_side_simulation', 'player_actions' => Actions::METHODS, 'session_actions' => self::SESSION_ACTIONS, 'personality_profiles' => Personality::PROFILES, 'control' => 'interactive_jsonl_and_individual_requests', 'limitations' => ['No real network connection, Xbox handshake, PlayerPreLoginEvent, rendering or resource-pack validation', 'Upstream simple physics does not implement swimming or ladders', 'Events are recorded only after explicit subscription', 'Packets and inbox are retained in runtime memory until explicit clearInbox or reset. Read them to persist a response before clearing. They are not automatically copied to the journal.']];
			case 'spawn': return $this->spawn(Actions::text($args, 'name'), $args);
			case 'list': return array_map(fn(Golem $g) => $this->state($g), array_values($this->golems));
			case 'reset':
				++$this->generation; $this->resetGolems(); return null;
			case 'wait': return $this->clock->wait(Actions::integer($args, 'ticks', 1, 0, 12000));
			case 'subscribe': return $this->subscribe(Actions::text($args, 'class'));
			case 'events': return $this->events(Actions::integer($args, 'after', 0, 0, PHP_INT_MAX));
			case 'assert': return $this->assertion($request, $args);
			case 'waitFor':
				return $this->clock->until(function() use ($request, $args) : mixed{
					try{ return $this->assertion($request, $args); }catch(\UnexpectedValueException){ return null; }
				}, Actions::integer($args, 'ticks', 100, 1, 12000), 'asserted plugin state');
			case 'script': return $this->script(Actions::text($args, 'file'), $args['params'] ?? []);
		}
		return Actions::execute($this, $this->golem(Actions::text($request, 'player')), $op, $args);
	}

	private function spawn(string $name, array $args) : Deferred{
		$traits = $args['traits'] ?? [];
		if(!is_array($traits)){ throw new \InvalidArgumentException('traits must be an object'); }
		$personality = new Personality(isset($args['profile']) ? Actions::text($args, 'profile') : 'balanced', Actions::integer($args, 'seed', random_int(0, 2147483647), 0, 2147483647), $traits);
		if(!preg_match('/^[A-Za-z][A-Za-z0-9_]{0,15}$/D', $name)){ throw new \InvalidArgumentException('Use a 1-16 character ASCII test name beginning with a letter'); }
		if(count($this->golems) >= 32){ throw new \LengthException('Explicit 32-golem limit reached'); }
		$server = $this->getServer();
		if(isset($this->golems[strtolower($name)]) || $server->getPlayerExact($name) !== null || $server->isOp($name) || $server->hasOfflinePlayerData($name)){
			throw new \InvalidArgumentException('Name belongs to an existing golem, real player, saved player or operator');
		}
		$generation = $this->generation;
		try{ $result = $this->factory->spawn($name); }
		catch(\Throwable $error){ $this->resetGolems(); throw $error; }
		$result->then(function(Golem $golem) use ($generation, $personality) : void{
			if($generation === $this->generation){
				$this->golems[strtolower($golem->name())] = $golem;
				$this->behaviours[strtolower($golem->name())] = new Behaviour($this, $golem, $personality);
				$this->observeBehaviour($golem->name(), ['action' => 'created', 'personality' => $personality->snapshot()]);
			}
		}, function(\Throwable $error) use ($generation) : void{
			if($generation === $this->generation){ ++$this->generation; $this->resetGolems(); }
		});
		return $result;
	}

	private function subscribe(string $class) : array{
		if(isset($this->subscriptions[$class])){ return ['class' => $class, 'already_subscribed' => true]; }
		if(count($this->subscriptions) >= 256){ throw new \LengthException('Explicit 256-subscription limit reached'); }
		if(!is_a($class, Event::class, true)){ throw new \InvalidArgumentException('class must be a loaded or autoloadable PocketMine event'); }
		$this->getServer()->getPluginManager()->registerEvent($class, function(Event $event) : void{
			$data = ['sequence' => ++$this->eventSequence, 'class' => $event::class, 'tick' => $this->getServer()->getTick(), 'cancelled' => $event instanceof Cancellable ? $event->isCancelled() : null, 'projection' => 'common_getters_only'];
			foreach(['getPlayer', 'getEntity', 'getBlock', 'getItem', 'getMessage', 'getDamage', 'getFrom', 'getTo'] as $getter){
				if(is_callable([$event, $getter]) && (new \ReflectionMethod($event, $getter))->getNumberOfRequiredParameters() === 0){
					$data[$getter] = View::value($event->$getter());
				}
			}
			$this->record('event', $data);
			$this->events[] = $data;
			if(count($this->events) > 4096){ array_shift($this->events); }
		}, EventPriority::MONITOR, $this, true);
		$this->subscriptions[$class] = true;
		return ['class' => $class, 'after' => $this->eventSequence];
	}

	private function events(int $after) : array{
		if($this->events !== [] && $after < $this->events[0]['sequence'] - 1){ throw new \OutOfBoundsException('Event cursor expired from memory. Complete subscribed event evidence remains in the journal.'); }
		return ['next' => $this->eventSequence, 'events' => array_values(array_filter($this->events, static fn(array $event) => $event['sequence'] > $after))];
	}

	private function assertion(array $request, array $args) : array{
		$actual = $this->state($this->golem(Actions::text($request, 'player')));
		$path = Actions::text($args, 'path');
		foreach(explode('.', $path) as $part){
			if(!is_array($actual) || !array_key_exists($part, $actual)){ throw new \UnexpectedValueException("Assertion path $path is absent"); }
			$actual = $actual[$part];
		}
		if(array_key_exists('equals', $args)){
			$passed = $actual === $args['equals'] || (is_numeric($actual) && is_numeric($args['equals']) && !is_string($actual) && !is_string($args['equals']) && (float) $actual === (float) $args['equals']);
		}elseif(array_key_exists('contains', $args)){
			$expected = $args['contains'];
			$passed = is_array($actual) ? in_array($expected, $actual, true) : (is_string($actual) && is_string($expected) && str_contains($actual, $expected));
		}else{ throw new \InvalidArgumentException('Assertion requires equals or contains'); }
		if(!$passed){ throw new \UnexpectedValueException('Assertion failed at ' . $path . '. Actual ' . json_encode($actual, JSON_THROW_ON_ERROR)); }
		return ['passed' => true, 'path' => $path, 'actual' => $actual];
	}

	private function script(string $file, mixed $params) : mixed{
		$root = realpath($this->getDataFolder() . 'scripts');
		$path = realpath($this->getDataFolder() . 'scripts/' . $file);
		if($path === false || $root === false || !str_starts_with($path, $root . DIRECTORY_SEPARATOR) || pathinfo($path, PATHINFO_EXTENSION) !== 'php'){
			throw new \InvalidArgumentException('Script must be a trusted PHP file inside the plugin data scripts directory');
		}
		if(filesize($path) > self::MAX_BYTES){ throw new \LengthException('Trusted script exceeds the explicit 1 MiB source limit'); }
		$this->record('script.source', ['file' => $file, 'sha256' => hash_file('sha256', $path), 'source_base64' => base64_encode(file_get_contents($path))]);
		$callable = require $path;
		if(!$callable instanceof \Closure){ throw new \InvalidArgumentException('Script must return a Closure accepting Agent and params'); }
		return $callable($this, $params);
	}

	private function record(string $kind, mixed $data) : void{
		$line = json_encode(['schema' => self::SCHEMA, 'boot' => $this->boot, 'sequence' => ++$this->sequence, 'time' => gmdate('Y-m-d\TH:i:s\Z'), 'elapsed_nano' => (string) (hrtime(true) - $this->started), 'tick' => $this->getServer()->getTick(), 'kind' => $kind, 'data' => $data], JSON_THROW_ON_ERROR | JSON_UNESCAPED_SLASHES) . "\n";
		if($this->journalBytes + strlen($line) > self::MAX_JOURNAL){ throw new \LengthException('Explicit 256 MiB simulation journal limit reached. Restart the runtime to create another journal.'); }
		if(file_put_contents($this->journal, $line, FILE_APPEND | LOCK_EX) !== strlen($line)){ throw new \RuntimeException('Simulation journal write failed'); }
		$this->journalBytes += strlen($line);
	}

	private function write(string $path, array $value) : void{
		$data = json_encode($value, JSON_THROW_ON_ERROR | JSON_UNESCAPED_SLASHES);
		if(file_put_contents($path . '.tmp', $data, LOCK_EX) !== strlen($data) || !rename($path . '.tmp', $path)){ throw new \RuntimeException('Mailbox write failed'); }
	}
}
