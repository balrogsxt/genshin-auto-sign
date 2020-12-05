<?php
$url = $_POST['url']??'';
$header = $_POST['header']??'[]';
$header = json_decode($header,true);
$data = $_POST['data']??'';
$ch = curl_init();
curl_setopt($ch, CURLOPT_URL, $url);
curl_setopt($ch, CURLOPT_RETURNTRANSFER, 1);
curl_setopt($ch, CURLOPT_HTTPHEADER, $header);
curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, FALSE);
curl_setopt($ch, CURLOPT_SSL_VERIFYHOST, FALSE);
curl_setopt($ch, CURLOPT_POST, 1);
curl_setopt($ch, CURLOPT_POSTFIELDS, $data);
$result = curl_exec($ch);
curl_close($ch);
$json = json_decode($result, true);
if(!$json){
    exit(json_encode(['retcode'=>-999,'message'=>'请求失败:'.$result],256));
}
exit($result);