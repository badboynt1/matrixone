-- @suit

-- @case
-- @desc:test for DELETE for multiple table
-- @label:bvt

DROP TABLE IF EXISTS t1;
DROP TABLE IF EXISTS t2;
DROP TABLE IF EXISTS t3;
DROP TABLE IF EXISTS t4;

CREATE TABLE t1(
                   id VARCHAR(10) PRIMARY KEY,
                   name VARCHAR(20)
);
CREATE TABLE t2(
                   id VARCHAR(10),
                   sex VARCHAR(4)
);
CREATE TABLE t3(
                   name VARCHAR(20),
                   score INT
);
CREATE TABLE t4(
                   id VARCHAR(10),
                   class VARCHAR(10)
);

INSERT INTO t1 VALUES('1','pet'), ('2','cat'), ('3','dog'), ('4','pig');
INSERT INTO t2 VALUES('1','M'), ('2','F'), ('4','F'), ('6',NULL);
INSERT INTO t3 VALUES('pet',101), ('dog',202), ('pig',303), ('dokey',505);
INSERT INTO t4 VALUES('2','c2'), ('4','c4'), ('6','c5'), ('7','c1');
SELECT * FROM t1;
DELETE t1 FROM t1 LEFT JOIN t2 ON t1.id = t2.id WHERE t2.id IS NULL;
SELECT * FROM t1;
SELECT * FROM t2;
DELETE t2 FROM t2 LEFT JOIN t1 ON t2.id = t1.id WHERE t2.sex = 'NULL';
SELECT * FROM t2;
DELETE t2 FROM t2 LEFT JOIN t1 ON t2.id = t1.id WHERE t2.sex = '';
SELECT * FROM t2;
SELECT * FROM t4;
DELETE t4 FROM t4 INNER JOIN t1 ON t4.id = t1.id WHERE t4.id BETWEEN 1 AND 3;
SELECT * FROM t4;
SELECT * FROM t3;
DELETE t3 FROM t3 RIGHT JOIN t1 ON t1.name = t3.name WHERE t3.score < 202;
SELECT * FROM t3;
DELETE t1,t2 FROM t1 LEFT JOIN t2 ON t1.id = t2.id WHERE t1.id = 25;
DELETE t1,t2 FROM t1,t2 WHERE t1.id = t2.id AND t1.id = 25;
DELETE t1 FROM t1,t2 WHERE t1.id = t2.id AND t1.id = 25;
DELETE t2 FROM t1,t2 WHERE t1.id = t2.id AND t1.id = 25;
DELETE t2 FROM t1,t2,t3 WHERE t1.id = t2.id OR t1.name = t3.name AND t1.id = 25;
DELETE t2 FROM t1,t2,t3,t4 WHERE t1.id = t2.id AND t1.id = t4.id AND t1.id = 25;
SELECT * FROM t2;

-- Whether support USING keyword
DELETE FROM t1 USING t1 LEFT JOIN t2 ON t1.id = t2.id WHERE t2.id IS NULL;
SELECT * FROM t1;
DELETE FROM t3 USING t1 LEFT JOIN t2 ON t1.id = t2.id WHERE t2.id IS NULL;

-- Alias
DELETE t1,t2 FROM table_name AS t1 LEFT JOIN table2_name AS t2 ON t1.id = t2.id WHERE table_name.id = 25;

DROP TABLE IF EXISTS t1;
DROP TABLE IF EXISTS t2;
DROP TABLE IF EXISTS t3;
DROP TABLE IF EXISTS t4;

CREATE TABLE t1(
                   t1_id INT NOT NULL PRIMARY KEY,
                   t1_o INT
);
CREATE TABLE t2(
                   t2_id INT NOT NULL PRIMARY KEY,
                   t1_id INT,
                   t2_o INT
);
CREATE TABLE t3(
                   t3_id INT NOT NULL PRIMARY KEY,
                   t2_id INT,
                   t3_o INT
);

INSERT INTO t1 VALUES(111,2), (222,3), (333,4);
INSERT INTO t2 VALUES(444,111,23), (555,222,34), (666,7,45);
INSERT INTO t3 VALUES(777,444,86), (888,666,56), (999,46,31);
SELECT * FROM t3;
DELETE t3 FROM t1 INNER JOIN t2 ON t1.t1_id = t2.t1_id INNER JOIN t3 ON t2.t2_id = t3.t3_id WHERE t1.t1_id = 222;
SELECT * FROM t3;
DELETE t1,t2,t3 FROM t1 INNER JOIN t2 ON t1.t1_id = t2.t1_id INNER JOIN t3 ON t2.t2_id = t3.t2_id WHERE t1.t1_id = 111;
SELECT * FROM t1;
SELECT * FROM t2;
SELECT * FROM t3;
-- Delete result is 0 rows affected on MO.
DELETE t1,t2,t3 FROM t1 LEFT JOIN t2 ON t1.t1_id = t2.t1_id RIGHT JOIN t3 ON t3.t2_id = t2.t2_id WHERE t2.t2_id = 555;
-- Delete result is 0 rows affected on MO.
DELETE t1,t2,t3 FROM t1 RIGHT JOIN t2 ON t1.t1_id = t2.t1_id RIGHT JOIN t3 ON t3.t2_id = t2.t2_id WHERE t2.t2_id = 555;

DELETE t1,t2,t3 FROM t1 RIGHT JOIN t2 ON t1.t1_id = t2.t1_id LEFT JOIN t3 ON t3.t2_id = t2.t2_id WHERE t2.t2_id = 555;
DELETE t1,t2,t3 FROM t1 LEFT JOIN t2 ON t1.t1_id = t2.t1_id LEFT JOIN t3 ON t3.t2_id = t2.t2_id WHERE t2.t2_id = 555;

DROP TABLE IF EXISTS t1;
DROP TABLE IF EXISTS t2;
DROP TABLE IF EXISTS t3;

CREATE TABLE t1(
                   a INT NOT NULL PRIMARY KEY,
                   b DATE,
                   c DATETIME
);
CREATE TABLE t2(
                   a INT NOT NULL PRIMARY KEY,
                   b INT
);
CREATE TABLE t3(
                   a VARCHAR(10) NOT NULL PRIMARY KEY,
                   b FLOAT
);
CREATE TABLE t4(
                   a CHAR(10) NOT NULL PRIMARY KEY,
                   b DOUBLE
);
CREATE TABLE t5(
                   a DOUBLE NOT NULL PRIMARY KEY,
                   b SMALLINT
);

INSERT INTO t1 VALUES (1,'2020-06-01','2023-04-05 13:16:46'), (2,'2019-04-19','2021-08-04 16:18:32'),(4,'2028-04-19','2021-08-04 16:18:32');
INSERT INTO t2 VALUES (1,897),(3,7984),(4,4894);
INSERT INTO t3 VALUES (3,0.001),(4,9.666),(5,46.321);
INSERT INTO t4 VALUES (3,123.11),(4,6464),(9,3641);
INSERT INTO t5 VALUES (4,153),(313,1561);

DELETE t1,t2,t3,t4,t5
FROM t1
INNER JOIN t2 ON t1.a = t2.a
INNER JOIN t3 ON t2.a = t3.a
INNER JOIN t4 ON t3.a = t4.a
INNER JOIN t5 ON t4.a = t5.a
WHERE t1.a = 1;

DELETE t1,t2 FROM t1,t2 WHERE t1.a = t2.a AND t1.a = 1;
SELECT * FROM t1;
SELECT * FROM t2;

DELETE t1,t5 FROM t1,t5 WHERE t1.a = t5.a AND t1.a = 4;
SELECT * FROM t1;
SELECT * FROM t5;

DROP TABLE IF EXISTS t1;
DROP TABLE IF EXISTS t2;
DROP TABLE IF EXISTS t3;
DROP TABLE IF EXISTS t4;
DROP TABLE IF EXISTS t5;
CREATE TABLE temp(
    id INT NOT NULL PRIMARY KEY
);
CREATE TABLE t1(
                   id INT NOT NULL PRIMARY KEY,
                   name VARCHAR(20)
);
INSERT INTO temp VALUES(1),(2),(3),(4),(5),(6),(7),(8),(9),(10);
INSERT INTO t1 VALUES(1,'likk'),(2,'fire'),(3,'wikky');

DELETE t1 FROM t1,(SELECT * FROM temp LIMIT 5) AS t WHERE t1.id = t.id AND t1.id > 2;
SELECT * FROM t1;

DROP TABLE IF EXISTS t1;
DROP TABLE IF EXISTS temp;

