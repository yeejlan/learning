<?php

function uuid() {
	$length = 16;
	$bytes = '';
	if(function_exists('openssl_random_pseudo_bytes')){
		$bytes = openssl_random_pseudo_bytes($length);
	}else{
	    for ($i = 1; $i <= $length; $i++) {
	        $bytes .= chr(mt_rand(0, 255));
	    }
	}
	return bin2hex($bytes);
}

for($i=0; $i<10; $i++){
	echo uuid(),PHP_EOL;
}