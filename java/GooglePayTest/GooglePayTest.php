<?php


function sign($content, $rsaPrivateKeyPem) {
        $priKey = file_get_contents($rsaPrivateKeyPem);
        $res = openssl_get_privatekey($priKey);
        openssl_sign($content, $sign, $res);
        openssl_free_key($res);
        $sign = base64_encode($sign);
        return $sign;
}


$content = 'my test data';
$rsaPrivateKeyPem = './rsa_private_key.pem';


$sign = sign($content, $rsaPrivateKeyPem);

//echo $sign, PHP_EOL;


function verifyGooglePurchase($base64PublicKey, $signedData, $signature) {
	$key = "-----BEGIN PUBLIC KEY-----\n" . wordwrap($base64PublicKey, 64, "\n", true) . "\n-----END PUBLIC KEY-----";
        $res = openssl_pkey_get_public($key);
        $ok = openssl_verify($signedData, base64_decode($signature), $res, 'sha1WithRSAEncryption');
        openssl_free_key($res);
        if($ok == 1) {
        	return true;
        }
        return false;
}



var_dump(verifyGooglePurchase(
"MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCoNVF/DoNOpY7vsWe4Nt+OUM4pZaBcG7ugTg0T347455ipLCH60YH1pF4N1AKDfGuG5HtGXCzOEH6KcLY0JKA5kEnzVwPXrovg1s9oAUp+f7+SDnymf2KbchraTHvtP718N3oxSaVl0JcxksfFIkQvS3nnQI6YoDPHpn9l1sBBlQIDAQAB",
		 "my test data",
		 "V18LTKZ3NjZyqkpYeqXByHXGHoZI+GpXSrZBEY43XtVcDdQype5x1RrIEbarjXi3jwBbcWStnyGS3858Bz5snLJI8SCa/cAROTKomWq2fqLMcMfedQga3uaS4BSYUtGOl14Rw3x0q9Z+DdNSKzIcq3mrIFON126psfLBaXUU6GM="

));		 