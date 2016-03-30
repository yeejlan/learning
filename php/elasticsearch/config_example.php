<?php

return array(

	'db'	=> array(
		'dsn'	=> 'mysql:host=localhost:3306;dbname=test',
		'user'	=> 'root',
		'pass'	=> 'mytestpasswd',
	),

	'sql'	=> 'select id as _id, id, title, content from posts where update_time > [datatime] and ctrate_timestamp > [timestamp]',

	'es'	=> 'http://localhost:9200',
	'index'	=> 'test',
	'type'	=> 'posts',

);