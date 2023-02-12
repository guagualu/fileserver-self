create database if not exists `fileserverself` DEFAULT CHARSET utf8mb4;

create table `user`(
    `id`  int  AUTO_INCREMENT  comment '用户id',
    `username` varchar(64) not null comment '用户昵称',
    `password` varchar(64) not null comment '用户密码',
    primary key(id),
    unique key(username)
)ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4;

create table `file`(
    `id` int AUTO_INCREMENT comment '文件id',
    `filehash` varchar(64) not null comment '文件hash',
    `filename` varchar(64) DEFAULT "" comment '文件名',
    `locateat` varchar(64) DEFAULT "" comment '文件存储地址',
    primary key(id),
    unique key(filehash)
)ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4;


后面来设外键? 和索引 必须要有个自增id 提高插入性能
create table `file_user`(
    `id` int AUTO_INCREMENT comment '假主键',   
    `username` varchar(64) not null comment '用户昵称',
    `filehash` varchar(64) not null comment '文件hash',
    `filename` varchar(64) not null comment '文件name',
    `status` int   comment '0代表被删 1代表正常',
    primary key(id),
    unique key(username),
    unique key(filehash)
)ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4;



todo :1、mysql索引的添加  2、分片文件的合并 3、rabbitmq的死信交换机 
添加索引 file_user 的username和filehash
create index uname_fhash on file_user(username,filehash);
添加索引 user 的username和password
create index name_pwd on `user`(`username`,`password`);