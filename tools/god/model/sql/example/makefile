#!/bin/bash

# 从脚本生成带缓存的模型
fromDDLWithCache:
	#god template clean
	god mysql ddl -src="./sql/*.sql" -dir="./sql/model/cache" -cache

fromDDLWithCacheAndIgnoreColumns:
	#god template clean
	god mysql ddl -src="./sql/*.sql" -dir="./sql/model/ignore_columns/cache" -cache -i 'gmt_create,create_time' -i 'gmt_modified,update_time'

fromDDLWithCacheAndDb:
	#god template clean
	god mysql ddl -src="./sql/*.sql" -dir="./sql/model/cache_db" -database="123nest" -cache

fromDDLWithoutCache:
	#god template clean;
	god mysql ddl -src="./sql/*.sql" -dir="./sql/model/nocache"


# 从数据源生成带缓存的模型
user=root
password=asdfasdf
datasource=127.0.0.1:3306
database=tutu

fromDataSource:
	#god template clean
	god mysql dsn -url="$(user):$(password)@tcp($(datasource))/$(database)" -table="*" -dir ./model/cache -c -style go_designer
