<?php

declare(strict_types=1);

namespace BedrockDebugProxy\GolemAgent;

use Golem\Golem;
use pocketmine\entity\Entity;
use pocketmine\inventory\Inventory;
use pocketmine\item\Item;
use pocketmine\math\Vector3;
use pocketmine\player\Player;
use pocketmine\network\mcpe\protocol\ClientboundPacket;
use pmmp\encoding\ByteBufferWriter;

final class View{
	public static function value(mixed $value, int $depth = 0) : mixed{
		if($depth > 12){
			throw new \LengthException('Result exceeds the explicit 12-level view limit. Use a focused script projection.');
		}
		if($value instanceof Golem){
			return ['name' => $value->name(), 'online' => $value->isOnline()];
		}
		if($value instanceof Item){
			return ['name' => $value->getName(), 'type_id' => $value->getTypeId(), 'count' => $value->getCount(), 'nbt' => (string) $value->getNamedTag()];
		}
		if($value instanceof Inventory){
			return ['size' => $value->getSize(), 'contents' => self::value($value->getContents(), $depth + 1)];
		}
		if($value instanceof Entity){
			return ['class' => $value::class, 'id' => $value->getId(), 'name' => $value instanceof Player ? $value->getName() : null, 'position' => self::value($value->getPosition()), 'alive' => $value->isAlive()];
		}
		if($value instanceof Vector3){
			return ['x' => $value->x, 'y' => $value->y, 'z' => $value->z];
		}
		if(is_string($value) && !mb_check_encoding($value, 'UTF-8')){
			return ['encoding' => 'base64', 'bytes' => strlen($value), 'data' => base64_encode($value)];
		}
		if(is_object($value)){
			if($value instanceof ClientboundPacket){
				$writer = new ByteBufferWriter();
				$value->encode($writer);
				return ['class' => $value::class, 'representation' => 'synthetic_server_packet_encoding', 'base64' => base64_encode($writer->getData()), 'public_fields' => self::value(get_object_vars($value), $depth + 1)];
			}
			return ['class' => $value::class, 'public_fields' => self::value(get_object_vars($value), $depth + 1)];
		}
		if(is_array($value)){
			if(count($value) > 16384){
				throw new \LengthException('Result exceeds the explicit 16384-item view limit. Use a focused script projection.');
			}
			foreach($value as $key => $item){
				$value[$key] = self::value($item, $depth + 1);
			}
		}
		return $value;
	}
}
