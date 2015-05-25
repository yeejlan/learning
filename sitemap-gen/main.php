<?php

$config = array(
	'db' => array(
		'host' => '127.0.0.1',
		'user' => 'root',
		'password' => '',
		'database' => 'sitemap',
	),
);

date_default_timezone_set('UTC');
main();

function main() {
	$shortopts  = "";
	$longopts  = array(
	    "generate::",     // generate sitemap
	    "crawl:",        // crawl the website
	    "initdb::",
	    "cleardatabase::",
	);
	$options = getopt($shortopts, $longopts);

	if(isset($options['crawl'])){
		global $website;
		$parts = parse_url($options['crawl']);
		$website = $options['crawl'];
		if(isset($parts['host'])) {
			$website = $parts['host'];
		}
		crawl($website);
	}elseif(isset($options['generate'])){
		generate_sitemap();
	}elseif(isset($options['cleardatabase'])){
		clear_database();
	}elseif(isset($options['initdb'])){
		initial_db();
	}else{
		echo 'This is a website crawler and sitemap generator', PHP_EOL;
		echo 'Please send one of those commands',PHP_EOL;
		print_r($longopts);
		echo PHP_EOL;
	}

}

function crawl($website) {
	echo 'Start with ',$website,PHP_EOL;
	$url = 'http://'.$website;
	add_one_url($url, 1, $url);
	crawl_with_level(1);
}

function get_total_page_by_level($level) {
	$link = connect_db();
	$query = "SELECT count(*) as cnt FROM pages WHERE level = {$level}";
	$result = mysqli_query($link, $query);
	if($result) {
		$row = $result->fetch_assoc();
		return $row['cnt'];
	}else{
		die(__FUNCTION__." error:[ level={$level} ] " . mysqli_error($link));
	}
}

function crawl_with_level($level) {
	$total_pages = get_total_page_by_level($level);
	if($total_pages<1) {
		die("ALL DONE with level: " .$level);
	}
	echo "crawl_with_level[ level={$level}, total_pages={$total_pages} ]",PHP_EOL;
	$link = connect_db();
	$query = "SELECT * FROM pages WHERE level = {$level}";
	$result = mysqli_query($link, $query);
	if($result) {
		while ($row = $result->fetch_assoc()) {
			scan_one_page($row['url'], $row['level']);
		}
	}else{
		die(__FUNCTION__." error:[ level={$level} ] " . mysqli_error($link));
	}

	crawl_with_level($level+1);
}

function scan_one_page($url, $level) {
	echo "scanning[ level={$level} ] : {$url}",PHP_EOL;
	list($code, $header, $text) = get_page_contents($url);
	if($code>=200 && $code<300){
		preg_match_all('|<a(.*)>(.*)</a>|U', $text, $matches, PREG_PATTERN_ORDER);
		$links = $matches[0];
		foreach($links as $link){
			$new_url = grap_url_from_link($link, $level);
			$new_url = normalize_url($new_url);
			if($new_url) {
				add_one_url($new_url, $level + 1, $url);
			}
		}
	}
	update_page_status($url, $code);
}

function normalize_url($url) {
	global $website;
	if(strpos($url, 'https') !== false){ //ignore https
		return false;
	}
	if(strpos($url, 'http') !== false){
		if(strpos($url, $website) === false) { //ignore external links
			return false;
		}
	}elseif($url[0] == '/'){
		$url = 'http://'.$website.$url;
	}else{
		return false;
	}

	$pos = strpos($url, '#');
	if($pos>0) {
		$url = substr($url, 0, $pos);
	}

	$url = rtrim($url, '/');
	return $url;
}

function grap_url_from_link($link, $level) {
	preg_match_all('|href="(.*)"|U', $link, $matches, PREG_PATTERN_ORDER);
	$urls = $matches[1];
	if(!isset($urls[0])){
		preg_match_all("|href='(.*)'|U", $link, $matches, PREG_PATTERN_ORDER);
	}
	$urls = $matches[1];
	if(isset($urls[0])) {
		return $urls[0];
	}
	return false;
}

function get_page_contents($url) {
	$options = array(
        CURLOPT_RETURNTRANSFER => true,     // return web page
        CURLOPT_HEADER         => true,    // return headers
        CURLOPT_FOLLOWLOCATION => true,     // follow redirects
        CURLOPT_ENCODING       => "",       // handle all encodings
        CURLOPT_USERAGENT      => "pwespider", // who am i
        CURLOPT_AUTOREFERER    => true,     // set referer on redirect
        CURLOPT_CONNECTTIMEOUT => 10,      // timeout on connect
        CURLOPT_TIMEOUT        => 10,      // timeout on response
        CURLOPT_MAXREDIRS      => 10,       // stop after 10 redirects
    );	
	$ch = curl_init( $url );
    curl_setopt_array( $ch, $options );
    $response = curl_exec( $ch );
    $err     = curl_errno( $ch );
    $errmsg  = curl_error( $ch );
    $info  = curl_getinfo( $ch );
    $http_code = $info['http_code'];
    $header_size = $info['header_size'];
    curl_close( $ch );
    if($err) {
    	echo $errmsg, PHP_EOL;
    }
    $header = substr($response, 0, $header_size);
	$body = substr($response, $header_size);
    return array($http_code, $header, $body);
}

function update_page_status($url, $http_code) {
	$link = connect_db();
	$url_crc = crc32($url);
	$query = "UPDATE pages SET page_code = {$http_code} WHERE url_crc = {$url_crc} AND URL = '{$url}' LIMIT 1";
	if (mysqli_query($link, $query) === TRUE) {
	    //pass
	}else{
		die(__FUNCTION__." error: [url = {$url}] ". mysqli_error($link));
	}
}

function add_one_url($url, $level, $referer_url) {
	$link = connect_db();
	$url_crc = crc32($url);
	$create_time = time();
	$query = "INSERT INTO pages(id, url, url_crc, level, referer_url, create_time)
		VALUES(NULL, '$url', {$url_crc}, {$level}, '{$referer_url}', {$create_time})";

	if (mysqli_query($link, $query) === TRUE) {
	    //pass
	}else{
		$error = mysqli_error($link);
		if(strpos($error, 'Duplicate entry') === FALSE) {
			die(__FUNCTION__." error: [url = {$url}, level={$level}] ". $error);
		}
	}
}

function clear_database() {
	$link = connect_db();
	$query = "TRUNCATE TABLE `pages`";
	if (mysqli_query($link, $query) === TRUE) {
	    //pass
	}else{
		die('TRUNCATE TABLE failed: '. mysqli_error($link));
	}	
}

function connect_db() {
	global $mysqli, $config;
	if(!$mysqli) {
		$mysqli = mysqli_connect($config['db']['host'], $config['db']['user'], $config['db']['password'], $config['db']['database']);
		if(mysqli_connect_errno()){
	    	die('Failed to connect to MySQL:' . mysqli_connect_error());
		}	
	}
	return $mysqli;
}

function initial_db() {
	$query = "
CREATE TABLE `pages` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `url` varchar(255) NOT NULL,
  `url_crc` bigint(10) NOT NULL,
  `level` int(11) NOT NULL,
  `page_code` int(11) NOT NULL DEFAULT '0',
  `referer_url` varchar(255) NOT NULL DEFAULT '-',
  `create_time` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `url` (`url`),
  KEY `url_crc` (`url_crc`)
) ENGINE=InnoDB  DEFAULT CHARSET=latin1;
";

	$link = connect_db();

	if (mysqli_query($link, $query) === TRUE) {
	    //pass
	}else{
		die(__FUNCTION__." error: ". mysqli_error($link));
	}	
	echo 'initial done.',PHP_EOL;
}

function generate_sitemap() {
	$urls_per_index = 20000;
	$total_page = get_total_page_200();
	$files_cnt = ceil($total_page/$urls_per_index);

	$link = connect_db();
	for($i=1; $i<=$files_cnt; $i++) {
		if($i == 1) {
			$idx = '';
		}else{
			$idx = $i;
		}
		$output_file = "sitemap{$idx}.xml";
		$records = array();

		$offset = ($i-1)*$urls_per_index;
		$query = "SELECT * FROM pages WHERE page_code = 200 order by id limit {$urls_per_index} offset {$offset}";
		$result = mysqli_query($link, $query);
		if($result) {
			while ($row = $result->fetch_assoc()) {
				$records[] = generate_one_record($row['url']);
			}
			write_sitemap($output_file, $records);
		}else{
			die(__FUNCTION__." error: " . mysqli_error($link));
		}		
	}
	write_sitemap_index($files_cnt);
	create_gz_files($files_cnt);
	echo "ALL DONE",PHP_EOL;

}

function create_gz_files($files_cnt) {
	if(PHP_OS != 'Linux') {
		echo 'Skip gzip',PHP_EOL;
		return;
	}
	for($i=1; $i<=$files_cnt; $i++) {
		if($i == 1) {
			$idx = '';
		}else{
			$idx = $i;
		}
		$sitemap_file = "sitemap{$idx}.xml";
		passthru("gzip -f < {$sitemap_file} > {$sitemap_file}.gz");
	}
	$sitemap_index_file = 'sitemap-index.xml';
	passthru("gzip -f < {$sitemap_index_file} > {$sitemap_index_file}.gz");

}

function write_sitemap($filename, $records) {
	$header = '<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
';
	$footer = '</urlset>';

	echo "Gennerating {$filename}...",PHP_EOL;
	$handle = fopen_or_die($filename, 'wa');
	fwrite_or_die($handle, $header);
	foreach($records as $record) {
		fwrite_or_die($handle, $record);
	}
	fwrite_or_die($handle, $footer);
	fclose($handle);
	echo "Done: {$filename}", PHP_EOL;	
}

function get_website_name() {
	$link = connect_db();
	$query = "SELECT url FROM pages order by id limit 1";
	$result = mysqli_query($link, $query);
	if($result) {
		$row = $result->fetch_assoc();
		$url = $row['url'];
		$parts = parse_url($url);
		return $parts['host'];
	}else{
		die(__FUNCTION__." error:[ level={$level} ] " . mysqli_error($link));
	}
}

function write_sitemap_index($files_cnt) {
	$website = get_website_name();

	$filename = 'sitemap-index.xml';
	$header = '<?xml version="1.0" encoding="UTF-8"?>
<sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
';
	$footer = '</sitemapindex>';	

	echo "Gennerating {$filename}...",PHP_EOL;
	$handle = fopen_or_die($filename, 'wa');
	fwrite_or_die($handle, $header);
	for($i=1; $i<=$files_cnt; $i++) {
		if($i == 1) {
			$idx = '';
		}else{
			$idx = $i;
		}
		$sitemap_file = "sitemap{$idx}.xml.gz";
		$loc = 'http://'.$website.'/'.$sitemap_file;
		$time = date('c');
		$record = "<sitemap><loc>{$loc}</loc><lastmod>{$time}</lastmod></sitemap>\r\n";
		fwrite_or_die($handle, $record);
	}
	fwrite_or_die($handle, $footer);
	fclose($handle);
	echo "Done: {$filename}", PHP_EOL;	
}

function fopen_or_die($filename, $mode) {
	$handle = fopen($filename, 'wa');
	if(!$handle) {
		die("Cannot open file: $filename");
	}
	return $handle;
}

function fwrite_or_die($handle, $content) {
	if (fwrite($handle, $content) === FALSE) {
        die("Cannot write to file");
    }
}

function generate_one_record($url) {
	$record = "<url><loc>{$url}</loc></url>\r\n";
	return $record;
}

function get_total_page_200() {
	$link = connect_db();
	$query = "SELECT count(*) as cnt FROM pages WHERE page_code = 200";
	$result = mysqli_query($link, $query);
	if($result) {
		$row = $result->fetch_assoc();
		return $row['cnt'];
	}else{
		die(__FUNCTION__." error: " . mysqli_error($link));
	}
}