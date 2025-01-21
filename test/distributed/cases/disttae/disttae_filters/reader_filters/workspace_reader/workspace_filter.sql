-- test mem filter apply on workspace entries
drop database if exists wf_db;
create database wf_db;
use wf_db;

create table t1(a int primary key);
insert into t1 values(1),(2),(3);

begin;
update t1 set a = 11 where a = 1;
update t1 set a = 22 where a = 2;
update t1 set a = 33 where a = 3;
insert into t1 values(44);
insert into t1 values(55);
select * from t1 where a > 44;
select * from t1 where a >= 55;
select * from t1 where a < 22;
select * from t1 where a <= 11;
select * from t1 where a >= 11 and a <= 22 order by a asc;
select * from t1 where a >= 11 and a < 22;
select * from t1 where a > 11 and a <= 22;
select * from t1 where a > 11 and a < 33;
delete from t1 where a = 44;
commit;
select * from t1 order by a asc;

create table t2(a varchar primary key);
insert into t2 values("1"),("2"),("3");
begin;
update t2 set a = "11" where a = "1";
update t2 set a = "22" where a = "2";
update t2 set a = "33" where a = "3";
insert into t2 values("44");
insert into t2 values("55");
delete from t2 where a = "44";
commit;
select * from t2 order by a asc;

create table t3(a int, b int, primary key (a,b));
insert into t3 values(1,1),(2, 2),(3,3);
begin;
update t3 set a = 11 where a = 1;
update t3 set a = 22 where a = 2;
update t3 set a = 33 where a = 3;
insert into t3 values(44, 4);
insert into t3 values(55, 5);
select * from t3 where a >= 11 and a <= 55 order by a asc;
select * from t3 where a in (11, 22) order by a asc;
delete from t3 where a = 44;
commit;
select * from t3 order by a, b asc;

drop table if exists t1;
drop table if exists t2;
drop table if exists t3;
drop database if exists wf_db;