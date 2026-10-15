<?php

declare(strict_types=1);

// No downloads or source execution. Package only this companion plugin.
if($argc !== 2 || ini_get('phar.readonly') === '1'){
	fwrite(STDERR, "Usage: php -d phar.readonly=0 dev/golem-agent/build.php OUTPUT.phar\n");
	exit(2);
}
$output = $argv[1];
if(file_exists($output)){ throw new RuntimeException('Output already exists. Choose another output path.'); }
$phar = new Phar($output);
$phar->startBuffering();
$phar->addFile(__DIR__ . '/plugin.yml', 'plugin.yml');
$phar->addFile(dirname(__DIR__, 2) . '/LICENSE', 'LICENSE');
$phar->addFile(__DIR__ . '/runtime-lock.json', 'runtime-lock.json');
$phar->addFile(__DIR__ . '/GOLEM-LICENSE.txt', 'GOLEM-LICENSE.txt');
foreach(['Agent', 'Actions', 'View', 'Personality', 'Behaviour'] as $class){
	$phar->addFile(__DIR__ . '/src/' . $class . '.php', 'src/' . $class . '.php');
}
$phar->setStub('<?php __HALT_COMPILER();');
$phar->setSignatureAlgorithm(Phar::SHA256);
$phar->stopBuffering();
echo "Built $output\n";
