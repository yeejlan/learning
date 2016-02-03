<?php

require 'ImageUploadService.php';

//upload image
$result = ImageUploadService::upload('test','e:/pic1.png');
var_dump($result);


//upload file with custom path
$result = ImageUploadService::upload('test','e:/pic1.zip', null, rand(1,10).'/'.time().'.zip');
var_dump($result);


//upload image with thumb
$resizeInfo = array('width'=>128, 'height'=>64);
$result = ImageUploadService::upload('test','e:/pic1.png', $resizeInfo);
var_dump($result);


//upload image with multiple thumbs
$resizeInfo = array(
	'media' => array('width'=>128, 'height'=>64),
	'homepage' => array('width'=>200, 'height'=>100),
	);

$result = ImageUploadService::upload('test','e:/pic1.png', $resizeInfo);
var_dump($result);
