<?php

$config = array(
	'ext' => array(
		'php',
		'phtml',
		'html',
		'css',
		'js',
		'less',
	),
);

$total = array(
	'cnt' => 0,
	'skip' => 0,
	'skipExt' => array(),
);

date_default_timezone_set('UTC');
main();


function main() {
	global $config;

	$shortopts  = "";
	$longopts  = array(
	    "check:",     // check dos style crlf
	    "replace:",        // replace dos style crlf to unix
	);
	$options = getopt($shortopts, $longopts);
	if(isset($options['check'])){
		checkAndReplaceDir($options['check'], $replace = false);
		echoStatistics();
	}elseif(isset($options['replace'])){
		checkAndReplaceDir($options['replace'], $replace = true);
		echoStatistics();
	}else{
		echo 'This is a crlf check and replace tool', PHP_EOL;

		echo 'Supported file extension :',PHP_EOL;
		print_r($config['ext']);
		echo PHP_EOL;

		echo 'Please send one of those commands',PHP_EOL;
		echo '  --check [DIR]',PHP_EOL;
		echo '  --replace [DIR]',PHP_EOL;
		echo PHP_EOL;
	}
}

function echoStatistics() {
	global $total;
	echo PHP_EOL;
	echo "Total handled: {$total['cnt']}, skipped: {$total['skip']}",PHP_EOL;
	echo "Skipped Extension: ", PHP_EOL;
	print_r(array_keys($total['skipExt']));
}

function checkAndReplaceDir($dir, $replace) {
	global $total;

	$dir = realpath($dir);
	if(is_file($dir)){
		checkAndReplaceOne($dir, $replace);
	}elseif(is_dir($dir)) {
		if ($dh = opendir($dir)) {
			while (($file = readdir($dh)) !== false) {
				$path = $dir.'/'.$file;
				if($file=="." || $file=="..") {
					//skip
					
				}elseif(is_dir($path)) { //dir

					checkAndReplaceDir($path, $replace);
				}else{ //file

					checkAndReplaceOne($path, $replace);
				}
			}
			closedir($dh);
		}else{
			die('Can not open dir: '.$dir);
		}
	}	
}


function checkAndReplaceOne($path, $replace) {
	global $config, $total;

	$total['cnt'] += 1;

	static $cnt = 0;
	$step = 100;
	$cnt++;
	if($cnt > $step) {
		printdot('.');
		$cnt = 0;
	}

	//check extension
	$info = pathinfo($path);
	$ext = $info['extension'];
	if(!in_array($ext, $config['ext'])) {
		$total['skip'] += 1;
		$total['skipExt'][$ext] = 1;
		return;
	}

	$text = file_get_contents($path);
	if($text === false) {
		printdot("Read file error: ".$path);
		die();
	}

	
	$pos = strpos($text, $find = "\r\n");
	if(false == $pos){
		return;
	}else{
		if(false == $replace) {
			printdot("[DOS] ".$path);
		}else{ //replace file
			$unixtxt = str_replace("\r\n", "\n", $text);
			$ret = file_put_contents($path, $unixtxt);
			if(false === $ret) {
				printdot("Write file error: ".$path);
				die();
			}
			printdot("[FIXED] ".$path);
		}
	}



}

function printdot($dot) {
	static $last = '.';

	if($dot == $last) {
		echo $dot;
	}else{
		echo PHP_EOL;
		echo $dot;
	}
	$last = $dot;
}