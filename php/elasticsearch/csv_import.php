<?php

set_time_limit(0);

define('DEFAULT_BUCK_SIZE', 1000);

main();

function main() {
    global $config;
    $shortopts  = "";
    $longopts  = array(
        "config:",
        "csv:",
        "separater:"
    );
    $options = getopt($shortopts, $longopts);
    
    if(!isset($options['config'])){
        help();
    }elseif(isset($options['csv'])){
        do_import($options);
        exit;
    }else{
        help();
    }
}

function help() {
        echo 'Elasticsearch csv data import tool.', PHP_EOL;
    
        echo PHP_EOL;

        echo 'Please send one of those commands',PHP_EOL;
        echo '  --config config_file, to specify the config file',PHP_EOL;
        echo '  --csv, the csv data file',PHP_EOL;
        echo '  --separater, data separater default "|", "tab" for "\t"',PHP_EOL;

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

    $csv_file = $options['csv'];
    if(!is_file($csv_file)){
        die('csv file not found: '. $csv_file);
    }
    if(!is_readable($csv_file)){
        die('csv file not readable: '. $csv_file);
    }

    $csv_handle = fopen($csv_file, 'r');
    if(!$csv_handle) {
        die('csv file open failed: '. $csv_file);
    }

    $total = 0;
    $time_begin = time();
    $cnt = 0;
    $data_formatted = '';
    $es_url = $config['es'].'/'.$config['index'].'/'.$config['type'].'/_bulk';
    while($line = fgets($csv_handle)){

        $row = format_one_line($line, $options['separater'], $config['csv_column']);

        if(isset($row['_id'])){
            $data_formatted .= '{ "index" : { "_id" : "'.$row['_id'].'" } }'."\n";
        }else{
            die('No _id found '. var_export($row, true));
        }
        unset($row['_id']);
        $data_formatted .= json_encode($row)."\n";
        $cnt ++ ;
        $total ++ ;
        if($cnt >= DEFAULT_BUCK_SIZE) {
            push_data_to_es($es_url, $data_formatted);
            $cnt = 0;
            $data_formatted = '';
        }

        if($total % ( 5*DEFAULT_BUCK_SIZE) == 0 ) {
            echo $total, ' ';
        }

    }
    if($data_formatted) {
        push_data_to_es($es_url, $data_formatted);
    }
    $now = date('c');
    echo PHP_EOL, "$now Total: {$total}, time cost: ", time()-$time_begin, ' second(s)', PHP_EOL;
}

function format_one_line($line, $separater, $column_mapping) {
    $line = trim($line);
    if(!$separater){
        $separater = '|';
    }
    if($separater == 'tab') {
        $separater =  "\t";
    }

    $data_arr = explode($separater, $line);
    $row = array();
    if(count($column_mapping) > count($data_arr)) {
        die('There are more columns in config than csv data, you may use the wrong seperator');
    }
    for($i=0; $i<count($column_mapping); $i++){
        $key = $column_mapping[$i];
        $value = $data_arr[$i];
        $row[$key] = $value;
    }
    return $row;
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

