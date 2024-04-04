-- pre
drop database if exists vecdb;
create database vecdb;
use vecdb;
drop table if exists vec_table;

-- standard
create table vec_table(a int, b vecf32(3), c vecf64(3));
desc vec_table;
insert into vec_table values(1, "[1,2,3]", "[4,5,6]");
select * from vec_table;

-- binary operators
select b+b from vec_table;
select b-b from vec_table;
select b*b from vec_table;
select b/b from vec_table;
select * from vec_table where b> "[1,2,3]";
select * from vec_table where b< "[1,2,3]";
select * from vec_table where b>= "[1,2,3]";
select * from vec_table where b<= "[1,2,3]";
select * from vec_table where b!= "[1,2,3]";
select * from vec_table where b= "[1,2,3]";
select * from vec_table where b= cast("[1,2,3]" as vecf32(3));
select b + "[1,2,3]" from vec_table;
select b + "[1,2]" from vec_table;
select b + "[1,2,3,4]" from vec_table;


-- cast
select cast("[1,2,3]" as vecf32(3));
select b + "[1,2,3]" from vec_table;
select b + sqrt(b) from vec_table;
select b + c from vec_table;

-- vector ops
select abs(b) from vec_table;
select abs(cast("[-1,-2,3]" as vecf32(3)));
select sqrt(b) from vec_table;
select summation(b) from vec_table;
select l1_norm(b) from vec_table;
select l2_norm(b) from vec_table;
select vector_dims(b) from vec_table;
select inner_product(b,"[1,2,3]") from vec_table;
select cosine_similarity(b,"[1,2,3]") from vec_table;
select l2_distance(b,"[1,2,3]") from vec_table;
select cosine_distance(b,"[1,2,3]") from vec_table;
select normalize_l2(b) from vec_table;

-- top K
select * FROM vec_table ORDER BY cosine_similarity(b, '[3,1,2]') LIMIT 5;
select * FROM vec_table ORDER BY l2_distance(b, '[3,1,2]') LIMIT 5;
select * FROM vec_table ORDER BY inner_product(b, '[3,1,2]') LIMIT 5;


-- throw error cases
select b + "[1,2,3" from vec_table;
select b + "1,2,3" from vec_table;
create table t2(a int, b vecf32(3) primary key);
create unique index t3 on vec_table(b);
create table t3(a int, b vecf32(65537));

-- throw error for Nan/Inf
select sqrt(cast("[1,2,-3]" as vecf32(3)));
select b/(cast("[1,2,0]" as vecf32(3))) from vec_table;

-- agg
select count(b) from vec_table;

-- insert test (more dim error)
create table t4(a int, b vecf32(5), c vecf64(5));
insert into t4 values(1, "[1,2,3,4,5]", "[1,2,3,4,5]");
insert into t4 values(1, "[1,2]", "[1,2]");
insert into t4 values(1, "[1,2,3,4,5,6]", "[1,2,3,4,5,6]");
select * from t4;

-- insert vector as binary
create table t5(a int, b vecf32(3));
insert into t5 values(1, decode('7e98b23e9e10383b2f41133f','hex'));
insert into t5 values(2, decode('0363733ff13e0b3f7aa39d3e','hex'));
insert into t5 values(3, decode('be1ac03e485d083ef6bc723f','hex'));

insert into t5 values(4, "[0,2,3]");

insert into t5 values(5, decode('05486c3f3ee2863e713d503dd58e8e3e7b88743f','hex')); -- this is float32[5]
insert into t5 values(6, decode('9be2123fcf92de3e','hex')); -- this is float32[2]

select * from t5;
select * from t5 where t5.b > "[0,0,0]";

-- output vector as binary (the output is little endian hex encoding)
select encode(b,'hex') from t5;

-- insert nulls
create table t6(a int, b vecf32(3));
insert into t6 values(1, null);
insert into t6 (a,b) values (1, '[1,2,3]'), (2, '[4,5,6]'), (3, '[2,1,1]'), (4, '[7,8,9]'), (5, '[0,0,0]'), (6, '[3,1,2]');
select * from t6;
update t6 set b = NULL;
select * from t6;

-- vector precision/scale check
create table t7(a int, b vecf32(3), c vecf32(5));
insert into t7 values(1, NULL,NULL);
insert into t7 values(2, "[0.8166459, 0.66616553, 0.4886152]", NULL);
insert into t7 values(3, "[0.1726299, 3.2908857, 30.433094]","[0.45052445, 2.1984527, 9.579752, 123.48039, 4635.894]");
insert into t7 values(4, "[8.560689, 6.790359, 821.9778]", "[0.46323407, 23.498016, 563.923, 56.076736, 8732.958]");
select * from t7;
select a, b + b, c + c from t7;
select a, b * b, c * c from t7;
select l2_norm(b), l2_norm(c) from t7;


-- insert, flush and select
insert into vec_table values(2, "[0,2,3]", "[4,4,6]");
insert into vec_table values(3, "[1,3,3]", "[4,1,6]");
-- @separator:table
select mo_ctl('dn', 'flush', 'vecdb.vec_table');
-- @separator:table
select mo_ctl('dn', 'flush', 'vecdb.t6');
select * from vec_table where b> "[1,2,3]";
select * from vec_table where b!= "[1,2,3]";
select * from vec_table where b= "[1,2,3]";

-- create table with PK or UK or No Key (https://github.com/matrixorigin/matrixone/issues/13038)
create table vec_table1(a int, b vecf32(3), c vecf64(3));
insert into vec_table1 values(1, "[1,2,3]", "[4,5,6]");
select * from vec_table1;
create table vec_table2(a int primary key, b vecf32(3), c vecf64(3));
insert into vec_table2 values(1, "[1,2,3]", "[4,5,6]");
select * from vec_table2;
create table vec_table3(a int unique key, b vecf32(3), c vecf64(3));
insert into vec_table3 values(1, "[1,2,3]", "[4,5,6]");
select * from vec_table3;

-- Scalar Null check
select summation(null);
select l1_norm(null);
select l2_norm(null);
select vector_dims(null);
select inner_product(null, "[1,2,3]");
select cosine_similarity(null, "[1,2,3]");
select l2_distance(null, "[1,2,3]");
select cosine_distance(null, "[1,2,3]");
select normalize_l2(null);
select cast(null as vecf32(3));
select cast(null as vecf64(3));

-- Precision issue for Cosine Similarity/Distance
create table t8(a int, b vecf32(3), c vecf32(5));
INSERT INTO `t8` VALUES (1,NULL,NULL);
INSERT INTO `t8` VALUES(2,'[0.8166459, 0.66616553, 0.4886152]',NULL);
INSERT INTO `t8` VALUES(3,'[0.1726299, 3.2908857, 30.433094]','[0.45052445, 2.1984527, 9.579752, 123.48039, 4635.894]');
INSERT INTO `t8` VALUES(4,'[8.560689, 6.790359, 821.9778]','[0.46323407, 23.498016, 563.923, 56.076736, 8732.958]');
select cosine_similarity(b,b), cosine_similarity(c,c) from t8;

create table t9(a int, b vecf64(3), c vecf64(5));
INSERT INTO `t9` VALUES (1,NULL,NULL);
INSERT INTO `t9` VALUES (2,'[0.8166459, 0.66616553, 0.4886152]',NULL);
INSERT INTO `t9` VALUES (3,'[8.5606893, 6.7903588, 821.977768]','[0.46323407, 23.49801546, 563.9229458, 56.07673508, 8732.9583881]');
INSERT INTO `t9` VALUES (4,'[0.9260021, 0.26637346, 0.06567037]','[0.45756745, 65.2996871, 321.623636, 3.60082066, 87.58445764]');
select cosine_similarity(b,b), cosine_similarity(c,c) from t9;

-- Sub Vector
create table t10(a int, b vecf32(3), c vecf64(3));
insert into t10 values(1, "[1,2.4,3]", "[4.1,5,6]");
insert into t10 values(2, "[3,4,5]", "[6,7.3,8]");
insert into t10 values(3, "[5,6,7]", "[8,9,10]");
select subvector(b,1) from t10;
select subvector(b,2) from t10;
select subvector(b,3) from t10;
select subvector(b,4) from t10;
select subvector(b,-1) from t10;
select subvector(b,-2) from t10;
select subvector(b,-3) from t10;
select subvector(b, 1, 1) from t10;
select subvector(b, 1, 2) from t10;
select subvector(b, 1, 3) from t10;
select subvector(b, 1, 4) from t10;
select subvector(b, -1, 1) from t10;
select subvector(b, -2, 1) from t10;
select subvector(b, -3, 1) from t10;
SELECT SUBVECTOR("[1,2,3]", 2);
SELECT SUBVECTOR("[1,2,3]",2,1);

-- Arithmetic Operators between Vector and Scalar
select b + 2 from t10;
select b - 2 from t10;
select b * 2 from t10;
select b / 2 from t10;
select 2 + b from t10;
select 2 - b from t10;
select 2 * b from t10;
select 2 / b from t10;
select b + 2.0 from t10;
select b - 2.0 from t10;
select b * 2.0 from t10;
select b / 2.0 from t10;
select 2.0 + b from t10;
select 2.0 - b from t10;
select 2.0 * b from t10;
select 2.0 / b from t10;
select cast("[1,2,3]" as vecf32(3)) + 2;
select cast("[1,2,3]" as vecf32(3)) - 2;
select cast("[1,2,3]" as vecf32(3)) * 2;
select cast("[1,2,3]" as vecf32(3)) / 2;
select 2 + cast("[1,2,3]" as vecf32(3));
select 2 - cast("[1,2,3]" as vecf32(3));
select 2 * cast("[1,2,3]" as vecf32(3));
select 2 / cast("[1,2,3]" as vecf32(3));
select cast("[1,2,3]" as vecf32(3)) + 2.0;
select cast("[1,2,3]" as vecf32(3)) - 2.0;
select cast("[1,2,3]" as vecf32(3)) * 2.0;
select cast("[1,2,3]" as vecf32(3)) / 2.0;
select 2.0 + cast("[1,2,3]" as vecf32(3));
select 2.0 - cast("[1,2,3]" as vecf32(3));
select 2.0 * cast("[1,2,3]" as vecf32(3));
select 2.0 / cast("[1,2,3]" as vecf32(3));
select cast("[1,2,3]" as vecf32(3)) / 0 ;
select 5 + (-1*cast("[1,2,3]" as vecf32(3)));

-- Distinct SQL
create table t11(a vecf32(2));
insert into t11 values('[1,0]');
insert into t11 values('[1,2]');
select distinct a from t11;
select distinct a,a from t11;

-- Except
select * from t8 except select * from t9;

-- Vector Default Non Null Value
create table vector_index_08(a int primary key, b vecf32(128),c int,key c_k(c));
create index idx01 using ivfflat on vector_index_08(b) lists=3 op_type "vector_l2_ops";
insert into vector_index_08 values(9774 ,"[1, 0, 1, 6, 6, 17, 47, 39, 2, 0, 1, 25, 27, 10, 56, 130, 18, 5, 2, 6, 15, 2, 19, 130, 42, 28, 1, 1, 2, 1, 0, 5, 0, 2, 4, 4, 31, 34, 44, 35, 9, 3, 8, 11, 33, 12, 61, 130, 130, 17, 0, 1, 6, 2, 9, 130, 111, 36, 0, 0, 11, 9, 1, 12, 2, 100, 130, 28, 7, 2, 6, 7, 9, 27, 130, 83, 5, 0, 1, 18, 130, 130, 84, 9, 0, 0, 2, 24, 111, 24, 0, 1, 37, 24, 2, 10, 12, 62, 33, 3, 0, 0, 0, 1, 3, 16, 106, 28, 0, 0, 0, 0, 17, 46, 85, 10, 0, 0, 1, 4, 11, 4, 2, 2, 9, 14, 8, 8]",3),(9775,"[0, 1, 1, 3, 0, 3, 46, 20, 1, 4, 17, 9, 1, 17, 108, 15, 0, 3, 37, 17, 6, 15, 116, 16, 6, 1, 4, 7, 7, 7, 9, 6, 0, 8, 10, 4, 26, 129, 27, 9, 0, 0, 5, 2, 11, 129, 129, 12, 103, 4, 0, 0, 2, 31, 129, 129, 94, 4, 0, 0, 0, 3, 13, 42, 0, 15, 38, 2, 70, 129, 1, 0, 5, 10, 40, 12, 74, 129, 6, 1, 129, 39, 6, 1, 2, 22, 9, 33, 122, 13, 0, 0, 0, 0, 5, 23, 4, 11, 9, 12, 45, 38, 1, 0, 0, 4, 36, 38, 57, 32, 0, 0, 82, 22, 9, 5, 13, 11, 3, 94, 35, 3, 0, 0, 0, 1, 16, 97]",5),(9776,"[10, 3, 8, 5, 48, 26, 5, 16, 17, 0, 0, 2, 132, 53, 1, 16, 112, 6, 0, 0, 7, 2, 1, 48, 48, 15, 18, 31, 3, 0, 0, 9, 6, 10, 19, 27, 50, 46, 17, 9, 18, 1, 4, 48, 132, 23, 3, 5, 132, 9, 4, 3, 11, 0, 2, 46, 84, 12, 10, 10, 1, 0, 12, 76, 26, 22, 16, 26, 35, 15, 3, 16, 15, 1, 51, 132, 125, 8, 1, 2, 132, 51, 67, 91, 8, 0, 0, 30, 126, 39, 32, 38, 4, 0, 1, 12, 24, 2, 2, 2, 4, 7, 2, 19, 93, 19, 70, 92, 2, 3, 1, 21, 36, 58, 132, 94, 0, 0, 0, 0, 21, 25, 57, 48, 1, 0, 0, 1]",3);
insert into vector_index_08 values(9777, "[16, 15, 0, 0, 5, 46, 5, 5, 4, 0, 0, 0, 28, 118, 12, 5, 75, 44, 5, 0, 6, 32, 6, 49, 41, 74, 9, 1, 0, 0, 0, 9, 1, 9, 16, 41, 71, 80, 3, 0, 0, 4, 3, 5, 51, 106, 11, 3, 112, 28, 13, 1, 4, 8, 3, 104, 118, 14, 1, 1, 0, 0, 0, 88, 3, 27, 46, 118, 108, 49, 2, 0, 1, 46, 118, 118, 27, 12, 0, 0, 33, 118, 118, 8, 0, 0, 0, 4, 118, 95, 40, 0, 0, 0, 1, 11, 27, 38, 12, 12, 18, 29, 3, 2, 13, 30, 94, 78, 30, 19, 9, 3, 31, 45, 70, 42, 15, 1, 3, 12, 14, 22, 16, 2, 3, 17, 24, 13]",4),(9778,"[41, 0, 0, 7, 1, 1, 20, 67, 9, 0, 0, 0, 0, 31, 120, 61, 25, 0, 0, 0, 0, 10, 120, 90, 32, 0, 0, 1, 13, 11, 22, 50, 4, 0, 2, 93, 40, 15, 37, 18, 12, 2, 2, 19, 8, 44, 120, 25, 120, 5, 0, 0, 0, 2, 48, 97, 102, 14, 3, 3, 11, 9, 34, 41, 0, 0, 4, 120, 56, 3, 4, 5, 6, 15, 37, 116, 28, 0, 0, 3, 120, 120, 24, 6, 2, 0, 1, 28, 53, 90, 51, 11, 11, 2, 12, 14, 8, 6, 4, 30, 9, 1, 4, 22, 25, 79, 120, 66, 5, 0, 0, 6, 42, 120, 91, 43, 15, 2, 4, 39, 12, 9, 9, 12, 15, 5, 24, 36]",4);
alter table vector_index_08 add column d vecf32(3) not null after c;
select * from vector_index_08;

-- post
drop database vecdb;