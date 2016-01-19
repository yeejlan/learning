<?php

//D:\games\Dota2\steamapps\workshop\content\570\

define('VPKPACKTOOL', 'D:/vpkpacktool/vpk.exe');

require_once 'KeyValues.php';

class D2maps {

	var $vpktool = VPKPACKTOOL;
	var $mapfile = '';
	var $base = './';

	public function D2maps($mapfile) {
		$this->mapfile = $mapfile;
	}

	public function _mkdir($base, $file) {
		$basepath = dirname($file);
		$pathname = $base.$basepath;
		if(!file_exists($pathname)) {
			mkdir($pathname, 644, true);
		}
	}

	public function _exit($message) {
		echo $message,PHP_EOL;
		die();
	}

	private function _removefile($file) {
		if(!file_exists($file)){
			return;
		}
		unlink($file);
	}

	public function _extractfile($file) {

		if(!file_exists($this->mapfile)){
			$this->_exit('Map not found: '. $this->mapfile);
		}

		$this->_mkdir($this->base, $file);
		$this->_removefile($file);

		$cmd = "{$this->vpktool} x {$this->mapfile} $file";
		$output = array();
		$return_var = 0;
		exec($cmd, $output, $return_var);

		if($return_var == 0 && count($output) == 1 && strpos($output[0], 'extracting') !== false) {
			return $this->base.$file;
		}
		print_r($output);
		$this->_exit("Read file error. file: {$file}, map: {$this->mapfile}");
	}

	public function readfile($file) {
		$path = $this->_extractfile($file);
		$content = file_get_contents($path);
		return $content;
	}

	public function readKv($file) {
		$content = $this->readfile($file);

		//UTF-16 (LE) detect
		if(bin2hex(substr($content,0,2)) == 'fffe'){
			$content = substr($content,2);
			$content = mb_convert_encoding($content, 'UTF-8', 'UCS-2LE');
		}

		$kv = new KeyValues(); 
		return $kv->load($content);
	}
}



$mapfile = 'E:/temp/455312245.vpk';
$map = new D2maps($mapfile);
$file = 'scripts/npc/npc_abilities_custom.txt';
$resultArr = $map->readKv($file);
echo json_encode($resultArr, JSON_PRETTY_PRINT);
