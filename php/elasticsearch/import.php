<?php

set_time_limit(0);

define('DEFAULT_BUCK_SIZE', 1000);

main();

function main() {
	global $config;
	$shortopts  = "";
	$longopts  = array(
		"config:",
	    "full::",     
	    "increment:",
	);
	$options = getopt($shortopts, $longopts);
	
	if(!isset($options['config'])){
		help();
	}elseif(isset($options['full'])){
		do_import($options);
		exit;
	}elseif(isset($options['increment'])){
		do_import($options, $options['increment']);
		exit;
	}else{
		help();
	}
}

function help() {
		echo 'Elasticsearch data import tool, need to have a timestamp or datetime colume to work.', PHP_EOL;
	
		echo PHP_EOL;

		echo 'Please send one of those commands',PHP_EOL;
		echo '  --config config_file, to specify the config file',PHP_EOL;
		echo '  --full, to make a full import',PHP_EOL;
		echo '  --increment interval_in_seconds, to make an increment import and specify the increment interval in seconds',PHP_EOL;

		echo PHP_EOL;

		echo 'Supported sql tag:',PHP_EOL;
		echo '  [timestamp]	', time(), PHP_EOL;
		echo '  [datetime]	', date('Y-m-d H:i:s'), PHP_EOL;

		echo PHP_EOL;
		exit;
}

function load_config($file) {
	if(!is_file($file)){
		die('Config file not found: '. $file);
	}
	if(!is_readable($file)){
		die('Config file not readable: '. $file);
	}
	$config = require $file;

	if(isset($config['timezone'])){
		date_default_timezone_set($config['timezone']);
	}

	echo 'Timezone: ', date_default_timezone_get(), PHP_EOL;

	return $config;
}


function do_import($options, $interval = null) {
	$config = load_config($options['config']);

	$dbh = new PDO($config['db']['dsn'], $config['db']['user'], $config['db']['pass']);
	$dbh->setAttribute(PDO::ATTR_DEFAULT_FETCH_MODE, PDO::FETCH_ASSOC);
	if(strpos(strtolower($config['db']['dsn']), 'mysql:') !== false){
		$dbh->exec('SET NAMES utf8');
	}


	$sql = $config['sql'];
	if($interval) {
		$timestamp = time() - $interval;
		$datetime = date('Y-m-d H:i:s', $timestamp);
		$sql = str_replace(array('[timestamp]','[datetime]'), array(':timestamp', ':datetime'), $sql);
		
	}

	$stmt = $dbh->prepare($sql);

	if(strpos($sql, ':timestamp')){
		$stmt->bindParam(':timestamp', $timestamp);
	}
	if(strpos($sql, ':datetime')){
		$stmt->bindParam(':datetime', $datetime);
	}

    if($stmt->execute() == false) {
        die('query db error ' . var_export($stmt->errorInfo(), true));
    }

    $total = 0;
    $time_begin = time();
    $cnt = 0;
    $data_formatted = '';
    $es_url = $config['es'].'/'.$config['index'].'/'.$config['type'].'/_bulk';
    while($row = $stmt->fetch(PDO::FETCH_ASSOC)){
 		if(isset($row['_id'])){
 			$data_formatted .= '{ "index" : { "_id" : "'.$row['_id'].'" } }'."\n";
 		}else{
 			die('No _id found '. var_export($row, true));
 		}
    	$data_formatted .= json_encode($row)."\n";
    	$cnt ++ ;
    	$total ++ ;
    	if($cnt >= DEFAULT_BUCK_SIZE) {
    		push_data_to_es($es_url, $data_formatted);
    		echo $total, ' ';
    		$cnt = 0;
    		$data_formatted = '';
    	}
    }
    if($data_formatted) {
    	push_data_to_es($es_url, $data_formatted);
    }
    $now = date('c');
    echo PHP_EOL, "$now Total: {$total}, time cost: ", time()-$time_begin, ' second(s)', PHP_EOL;
}

function push_data_to_es($url, $data_formatted) {

		$ch = curl_init();
		curl_setopt($ch, CURLOPT_URL, $url);
		curl_setopt($ch, CURLOPT_RETURNTRANSFER, 1 );
		curl_setopt($ch, CURLOPT_POST, 1 );
		curl_setopt($ch, CURLOPT_POSTFIELDS, $data_formatted);
		$result = curl_exec($ch);

		$curlErrMsg = '';
		$curlInfo = curl_getinfo($ch);
		if($result === false) {
			$curlErrMsg = 'Curl error: ('.curl_error($ch).')'. var_export($curlInfo, true);
		} else {
			if($curlInfo['http_code'] >=300 || $curlInfo['http_code'] <200) {
				$curlErrMsg = 'Http error: ['.$curlInfo['http_code'].']('.curl_error($ch).')'. var_export($curlInfo, true);
			}
		}
		curl_close($ch);
		
		if($curlErrMsg !== '') {
			die($result.PHP_EOL.$curlErrMsg);
		}
		//echo $result;exit;

}

