<?php
ini_set('memory_limit', '256M');

#error_reporting(E_ALL & ~E_NOTICE);
#ini_set('display_errors', 'On');

define('IMG_SERVICE_KEY', 'the_secret_key_280dc090a1a0833151632a441548d51d');
define('UPLOAD_PATH_ROOT', '/web/service.example.com/upload/');
define('IMAGE_URL_ROOT', 'http://image.example.com/');

$siteList = array(
	'test',
    'avatar',
    'game',
    'portal',
    'file',
);

$allowExt = array(
	'bmp',
	'gif', 
	'jpg', 
	'jpeg',
	'png',
	'flv',
	'mp4',
	'webm',
	'pdf',
	'doc',
	'xls',
	'docx',
	'xlsx',
	'zip',
	'rar',
	'ico',
	);

$site = trim($_GET['site']);
$name = base64_decode($_GET['name']);
$rand = trim($_GET['rand']);
$key = trim($_GET['key']);
$size = trim($_GET['size']);
$resize = json_decode(base64_decode($_GET['resize']), true);
$customCdnPath = base64_decode($_GET['cusCdnName']);

if($key != hash_hmac('md5', $site.$size.$name.$rand, $IMG_SERVICE_KEY)){
	die('bad request.');
}

if(!in_array($site, $siteList)){
	die('bad site config.['.$site.'] not found in siteList');
}

define('UPLOAD_PATH',  UPLOAD_PATH_ROOT.$site);
define('IMAGE_URL_BASE', IMAGE_URL_ROOT.$site);

if(!file_exists(UPLOAD_PATH)){
	mkdirOrDie(UPLOAD_PATH);
}

//check extention name
$filenameArr = explode('.', basename($name));
$ext = strtolower($filenameArr[count($filenameArr)-1]);
if(!in_array($ext, $allowExt)){
	die('bad file type.');
}

//get post data
$postStr = '';
$fp = fopen('php://input','r');
while (!feof($fp)){
	$postStr .= fgets($fp, 4096);
}
if(strlen($postStr) != $size){
	die('size not match.');
}

//write a temp file
$tempFile = getTempFileName($name);
$writeSize = file_put_contents($tempFile, $postStr);
if($writeSize != $size){
	@unlink($tempFile);
	die('write size not match.');
}

//resize image
$resizedImgArr = array();
$resizedImgArr2 = array();
if(is_array($resize)){
	if(isset($resize['width']) && isset($resize['height'])){
		$resize2[0] = $resize;
		unset($resize);
		$resize = $resize2;
	}
}

//create final file name
if($customCdnPath==''){
	$imgFilePath = getUniqueFileName($tempFile);
}else{
	$imgFilePath = getCustomCdnFileName($customCdnPath);
}

if(is_array($resize) && count($resize) > 0){
	foreach($resize as $idx => $value){
		if(intval($value['width'])<10 || intval($value['height'])<10 || intval($value['width'])>1024 || intval($value['height'])>1024){
			die('bad resize params.');
		}
	}
}

if(is_array($resize) && count($resize) > 0){
	foreach($resize as $idx => $value){
		$uniqueFile = getFileNameForThumbImg($imgFilePath, $value['width'], $value['height']);
		$thumbImg = UPLOAD_PATH.$uniqueFile;
		if(!createThumbImg($tempFile, $thumbImg, $value['width'], $value['height'])){
			@unlink($tempFile);
			foreach($resizedImgArr as $img){
				@unlink(UPLOAD_PATH.$img);
			}
			die('create thumbnail error.');
		}
		$resizedImgArr[$idx] = $thumbImg;
		$resizedImgArr2[$idx] = IMAGE_URL_BASE.$uniqueFile;
	}
}
//move temp file to upload path
$bigImg = UPLOAD_PATH.$imgFilePath;

if(!copy($tempFile, $bigImg)) {
	@unlink($tempFile);
	foreach($resizedImgArr as $img){
		@unlink(UPLOAD_PATH.$img);
	}
	die('copy image error.');
}

@unlink($tempFile);


echo 'success|';
$resultArr = array(
	'image' => IMAGE_URL_BASE.$imgFilePath,
	'thumb' => $resizedImgArr2,
	);
echo json_encode($resultArr);

exit;

function getCustomCdnFileName($customCdnPath, $basePath= UPLOAD_PATH){
	$customCdnPath = str_replace("\\", '/', $customCdnPath);
	if(substr($customCdnPath,0,1)=='/'){
		$customCdnPath = substr($customCdnPath,1);
	}
	$pathArr = explode('/', $customCdnPath);
	$dir = $basePath;
	for($i=0;$i<count($pathArr)-1;$i++){
		$dir = $dir.'/'.$pathArr[$i];
		if(!file_exists($dir)){
			mkdirOrDie($dir);
		}
	}
	$fullPath = '/'.$customCdnPath;
	return $fullPath;
}

function getUniqueFileName($srcImage, $basePath= UPLOAD_PATH){

		$filenameArr = explode('.', basename($srcImage));
		$ext = strtolower($filenameArr[count($filenameArr)-1]);

		$uuid = uuid();
		$dira = $basePath.'/'.substr($uuid, 0, 2);
		$dirb = $dira.'/'.substr($uuid, 2, 2);
		$filename = substr($uuid, 4);
		clearstatcache();
		if(!file_exists($dira)){
			mkdirOrDie($dira);
		}
		if(!file_exists($dirb)){
			mkdirOrDie($dirb);
		}
		$fullPath = $dirb.'/'.$filename.time().".$ext";
		return substr($fullPath, strlen($basePath));
}

function getFileNameForThumbImg($srcImage, $width, $height){
		$filenameArr = explode('.', basename($srcImage));
		$ext = '.'.strtolower($filenameArr[count($filenameArr)-1]);
		$newExt = '_'.$width.'x'.$height.$ext;
		$filename = str_replace($ext, $newExt, $srcImage);
		return $filename;
}

function getTempFileName($srcImage){
		$rand = rand(1, 10000).rand(1, 1000).rand(1, 100);
		list($usec, $sec) = explode(" ", microtime());
		$mtime = ((float)$usec + (float)$sec);
		$uid = uniqid();
		$filenameArr = explode('.', basename($srcImage));
		$ext = strtolower($filenameArr[count($filenameArr)-1]);
		return '/tmp/img_'.md5($rand.$mtime.$uid.$srcImage).time().".$ext";
}

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

function createThumbImg($imgsrc, $thumbpath, $thumbw= 130, $thumbh = 86){

		$imsrc = @imagecreatefromstring(file_get_contents($imgsrc));
		if(!$imsrc){
			return false;
		}
		
		$orgratio = $thumbw/$thumbh;

		$width = imagesx($imsrc);
		$height = imagesy($imsrc);

		$newratio = $width/$height;
		if($orgratio>$newratio){
			$newh = $thumbh*$width/$thumbw;
			$deltah = ($height-$newh)/3;
			$srcx = 0;
			$srcy = 0 + floor($deltah);
			$srcw = $width;
			$srch = $newh;
		}else{
			$neww = $thumbw*$height/$thumbh;
			$deltaw = ($width-$neww)/2;
			$srcx = 0 + floor($deltaw);
			$srcy = 0;
			$srcw = $neww;
			$srch = $height;
		}
		
		$imdst = @imagecreatetruecolor($thumbw, $thumbh);
		if(!$imdst || !$imsrc)
			return false;
		if(!@imagecopyresampled($imdst,$imsrc,0,0,$srcx,$srcy,$thumbw,$thumbh,$srcw,$srch))
			return false;
		if(!@imagejpeg($imdst,$thumbpath,80))
			return false;
		imagedestroy($imdst);
		imagedestroy($imsrc);
		return true;
}

function mkdirOrDie($path) {
	if(!mkdir($path)) {
		die("create dir failed."); //{$path}
	}
}

