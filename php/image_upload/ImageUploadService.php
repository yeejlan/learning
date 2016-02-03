<?php

define('IMG_SERVICE_KEY', 'the_secret_key_280dc090a1a0833151632a441548d51d');
define('IMG_SERVICE_URL', 'http://service.example.com/upload.php');

class ImageUploadService {

	/*
	*	//upload image
	*	$result = ImageUploadService::upload('test','e:/pic1.png');
	*
	*	//upload file with custom path
	*	$result = ImageUploadService::upload('test','e:/pic1.zip', null, rand(1,10).'/'.time().'.zip');
	*
	*	//upload image with thumb
	*	$resizeInfo = array('width'=>128, 'height'=>64);
	*	$result = ImageUploadService::upload('test','e:/pic1.png', $resizeInfo);	
	*
	*	//upload image with multiple thumbnails
	*	$resizeInfo = array(
	*	'media' => array('width'=>128, 'height'=>64),
	*	'homepage' => array('width'=>200, 'height'=>100),
	*	);
	*	$result = ImageUploadService::upload('test','e:/pic1.png', $resizeInfo);
	*
	*/
	static function upload($site, $srcImage, $resizeInfo = array(), $customCdnPath=''){
		if($customCdnPath!='' && count($resizeInfo)>0){
			return array('result' => 'failed', 'data' =>'Can\'t use $customCdnPath with $resizeInfo together' );
		}
		$fileContent = file_get_contents($srcImage);
		$size = strlen($fileContent);
		$name = basename($srcImage);
		
		$resize = urlencode(base64_encode(json_encode($resizeInfo)));
		$rand = rand(1, 10000);
		$key = hash_hmac('md5', $site.$size.$name.$rand, $IMG_SERVICE_KEY);
		$name = urlencode(base64_encode($name));
		$cusCdnName = urlencode(base64_encode($customCdnPath));

		$remoteUrl = IMG_SERVICE_URL."?site={$site}&size={$size}&name={$name}&rand={$rand}&key={$key}&resize={$resize}&cusCdnName={$cusCdnName}";

		$ch = curl_init();
		curl_setopt($ch, CURLOPT_URL,  $remoteUrl);
		curl_setopt($ch, CURLOPT_RETURNTRANSFER, 1 );
		curl_setopt($ch, CURLOPT_POST, 1 );
		curl_setopt($ch, CURLOPT_POSTFIELDS,  $fileContent); 
		curl_setopt($ch, CURLOPT_HTTPHEADER, array('Content-Type: binary/octet-stream')); 
		$result = curl_exec($ch);

		if(substr($result,0,strlen('success')) != 'success'){
			return array('result' => 'failed', 'data' => $result);
		}
		$result = substr($result, strlen('success')+1);
		return array('result' => 'success', 'data' => json_decode($result, true));
	}

	public static function createThumbImg($imgsrc, $thumbpath, $thumbw= 130, $thumbh = 86){
			$imgbuffer = file_get_contents($imgsrc);
			if(!$imgbuffer){
				return false;
			}
			$imsrc = imagecreatefromstring($imgbuffer);
			if(!$imsrc){
				return false;
			}
			
			$width = imagesx($imsrc);
			$height = imagesy($imsrc);
			if($width<1 || $height<1){
				return false;
			}
			
			$orgratio = $thumbw/$thumbh;
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
			
			$imdst = imagecreatetruecolor($thumbw, $thumbh);
			if(!$imdst || !$imsrc)
				return false;
			if(!imagecopyresampled($imdst,$imsrc,0,0,$srcx,$srcy,$thumbw,$thumbh,$srcw,$srch))
				return false;
			if(!imagejpeg($imdst,$thumbpath,80))
				return false;
			imagedestroy($imdst);
			imagedestroy($imsrc);
			return true;
	}
	
	static function getTempFileName($srcImage){
        $rand = rand(1, 10000).rand(1, 1000).rand(1, 100);
        list($usec, $sec) = explode(" ", microtime());
        $mtime = ((float)$usec + (float)$sec);
        $uid = uniqid();
        $filenameArr = explode('.', basename($srcImage));
        $ext = strtolower($filenameArr[count($filenameArr)-1]);
        return 'img_'.md5($rand.$mtime.$uid.$srcImage).time().".$ext";
    }


	/**
	 * 剪切图像的某一块区域并缩放
	 * @param string $srcFile  源图片路径
	 * @param string $destFile 目标文件路径
	 * @param integer $cropimage_x 剪切点的x坐标
	 * @param integer $cropimage_x 剪切点的y坐标
	 * @param integer $cropimage_w 剪切的宽度
	 * @param integer $cropimage_h 剪切的高度
	 * @param integer $targ_w 缩放后的宽度
	 * @param integer $targ_h 缩放后的高度
	 *
	 */
	public function imagickCropImage($srcFile, $destFile, $cropimage_x, $cropimage_y, $targ_w, $targ_h, $cropimage_w, $cropimage_h){
		if($cropimage_x < 0 || $cropimage_y < 0 || $targ_w <= 0 || $targ_h <= 0 || $cropimage_w <= 0 || $cropimage_h <= 0) {
			return false;
		}
		try {
			$src = new Imagick($srcFile);
		} catch(Exception $e) {
			return false;
		}
		$image_format = strtolower($src->getImageFormat());
		if($image_format != 'jpeg' && $image_format != 'gif' && $image_format != 'png' && $image_format != 'jpg' && $image_format != 'bmp') {
			return false;
		}

		/* 如果是 jpg jpeg png bmp */
		if($image_format != 'gif'){
			$dest = $src;
			$dest->cropImage ($cropimage_w, $cropimage_h, $cropimage_x, $cropimage_y);
			$dest->adaptiveResizeImage($targ_w, $targ_h);
			
			$dest->writeImage($destFile);
			$dest->clear();
		} else {
			/* gif需要以帧一帧的处理 */
			$dest = new Imagick();
			$color_transparent = new ImagickPixel("transparent"); //透明色
			foreach($src as $img){
				$page = $img->getImagePage();
				$tmp = new Imagick();
				$tmp->newImage($page['width'], $page['height'], $color_transparent, 'gif');
				$tmp->compositeImage($img, Imagick::COMPOSITE_OVER, $page['x'], $page['y']);
				
				$tmp->cropImage ($cropimage_w, $cropimage_h, $cropimage_x, $cropimage_y);
				$tmp->adaptiveResizeImage($targ_w, $targ_h);
				
				if(isset($watermask) && $watermask)
					$tmp->compositeImage($water, Imagick::COMPOSITE_OVER, $tmp->getImageWidth() - $water_w, $tmp->getImageHeight() - $water_h);
				$dest->addImage($tmp);
				$dest->setImagePage($tmp->getImageWidth(), $tmp->getImageHeight(), 0, 0);
				$dest->setImageDelay($img->getImageDelay());
				$dest->setImageDispose($img->getImageDispose());
			}
			$dest->coalesceImages();
			$dest->writeImages($destFile, true);
			$dest->clear();
		}
		return true;
	}

}