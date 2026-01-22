-- write db base here 
CREATE TABLE EMPLOYEE (
    Fname VARCHAR(15) NOT NULL,
    Minit CHAR,
    Lname VARCHAR(15) NOT NULL,
    Ssn CHAR(9) NOT NULL,
    Bdate DATE,
    Address VARCHAR(30),
    Sex CHAR,
    Salary DECIMAL(10, 2),
    Super_ssn CHAR(9),
    Dno INT NOT NULL,
    PRIMARY KEY (Ssn)
);
CREATE TABLE DEPARTMENT (
    Dname VARCHAR(15) NOT NULL,
    Dnumber INT NOT NULL,
    Mgr_ssn CHAR(9) NOT NULL,
    Mgr_start_date DATE,
    PRIMARY KEY(Dnumber),
    UNIQUE(Dname),
    FOREIGN KEY (Mgr_ssn) REFERENCES EMPLOYEE(Ssn)
);
CREATE TABLE DEPT_LOCATIONS (
    Dnumber INT NOT NULL,
    Dlocation VARCHAR(15) NOT NULL,
    PRIMARY KEY (Dnumber, Dlocation),
    FOREIGN KEY (Dnumber) REFERENCES DEPARTMENT(Dnumber)
);
CREATE TABLE PROJECT (
    Pname VARCHAR(15) NOT NULL,
    Pnumber INT NOT NULL,
    Plocation VARCHAR(15),
    Dnum INT NOT NULL,
    PRIMARY KEY (Pnumber),
    UNIQUE (Pname),
    FOREIGN KEY(Dnum) REFERENCES DEPARTMENT(Dnumber)
);
CREATE TABLE WORKS_ON (
    Essn CHAR(9) NOT NULL,
    Pno INT NOT NULL,
    Hours DECIMAL(3, 1) NOT NULL,
    PRIMARY KEY (Essn, Pno),
    FOREIGN KEY (Essn) REFERENCES EMPLOYEE(Ssn),
    FOREIGN KEY (Pno) REFERENCES PROJECT(Pnumber)
);
CREATE TABLE DEPENDENT (
    Essn CHAR(9) NOT NULL,
    Dependent_name VARCHAR(15) NOT NULL,
    Sex CHAR,
    Bdate DATE,
    Relationship VARCHAR(8),
    PRIMARY KEY (Essn, Dependent_name),
    FOREIGN KEY (Essn) REFERENCES EMPLOYEE(Ssn)
);


INSERT INTO EMPLOYEE (Fname, Lname, Dno, Ssn)
VALUES ('Pedram', 'Pouya', '12', '1');

INSERT INTO EMPLOYEE (Fname, Lname, Ssn, Bdate, Sex, Salary, Super_ssn, Dno)
VALUES ('John', 'Smith', '123456789', DATE'1965-01-09', 'M', 10000, '1', 5);

INSERT INTO EMPLOYEE (Fname, Lname, Ssn, Bdate, Sex, Salary, Super_ssn, Dno)
VALUES ('Franklin', 'Wong', '333445555', DATE'1955-12-08', 'M', 20000, '123456789', 5);

INSERT INTO EMPLOYEE (Fname, Lname, Ssn, Bdate, Sex, Salary, Super_ssn, Dno)
VALUES ('Alicia', 'Zelaya', '999887777', DATE'1968-07-19', 'F', 30000, '123456789', 4);

INSERT INTO EMPLOYEE (Fname, Lname, Ssn, Bdate, Sex, Salary, Super_ssn, Dno)
VALUES ('Jennifer', 'Wallace', '987654321', DATE'1941-06-20', 'F', 40000, '123456789', 4);

INSERT INTO EMPLOYEE (Fname, Lname, Ssn, Bdate, Sex, Salary, Super_ssn, Dno)
VALUES ('Ramesh', 'Narayan', '666884444', DATE'1962-09-15', 'F', 50000, '123456789', 5);

INSERT INTO EMPLOYEE (Fname, Lname, Ssn, Bdate, Sex, Salary, Super_ssn, Dno)
VALUES ('Joyce', 'English', '453453453', DATE'1972-07-31', 'M', 60000, '453453453', 5);

INSERT INTO EMPLOYEE (Fname, Lname, Ssn, Bdate, Sex, Salary, Super_ssn, Dno)
VALUES ('Ahmad', 'Jabbar', '987987987', DATE'1969-03-29', 'M', 70000, '666884444', 4);

INSERT INTO EMPLOYEE (Fname, Lname, Ssn, Bdate, Sex, Salary, Super_ssn, Dno)
VALUES ('James', 'Borg', '888665555', DATE'1937-11-10', 'M', 70000, '666884444', 1);

INSERT INTO EMPLOYEE (Fname, Lname, Ssn, Bdate, Sex, Salary, Super_ssn, Dno)
VALUES ('Bahram', 'Joya', '123321159', DATE'1937-11-10', 'M', 70000, NULL, 5);

-- department 

INSERT INTO DEPARTMENT
VALUES ('Research', 5, 333445555, NULL);

INSERT INTO DEPARTMENT
VALUES ('Administration', 4, 987654321, NULL);

INSERT INTO DEPARTMENT
VALUES ('Headquarters', 1, 888665555, NULL);


-- dependent
insert INTO DEPENDENT
VALUES('333445555', 'Alice', 'F', DATE'1986-04-05', 'Daughter');

insert INTO DEPENDENT
VALUES('333445555', 'Theodore', 'M', DATE'1983-10-25', 'Son');

insert INTO DEPENDENT
VALUES('333445555', 'Joy', 'F', DATE'1958-05-03', 'Spouse');

insert INTO DEPENDENT
VALUES('987654321', 'Abner', 'M', DATE'1942-02-28', 'Spouse');

insert INTO DEPENDENT
VALUES('123456789', 'Michael', 'M', DATE'1988-01-04', 'Son');

insert INTO DEPENDENT
VALUES('123456789', 'Alice', 'F', DATE'1988-12-30', 'Daughter');

insert INTO DEPENDENT
VALUES('123456789', 'Elizabeth', 'F', DATE'1967-05-05', 'Spouse');

--dept location
INSERT INTO DEPT_LOCATIONS
VALUES(1, 'Houston');

INSERT INTO DEPT_LOCATIONS
VALUES(4, 'Stafford');

INSERT INTO DEPT_LOCATIONS
VALUES(5, 'Bellaire');

INSERT INTO DEPT_LOCATIONS
VALUES(5, 'Sugarland');

INSERT INTO DEPT_LOCATIONS
VALUES(5, 'Houston');

-- project 
-- pname, pnumebr, plocation , dnum

INSERT INTO PROJECT
VALUES('ProductX', 1, 'Bellaire', 5);

INSERT INTO PROJECT
VALUES('ProductY', 2, 'Sugarland', 5);

INSERT INTO PROJECT
VALUES('ProductZ', 3, 'Houston', 5);

INSERT INTO PROJECT
VALUES('Computerization', 10, 'Stafford', 4);

INSERT INTO PROJECT
VALUES('Reorganization', 20, 'Houston', 1);

INSERT INTO PROJECT
VALUES('Newbenefits', 30, 'Stafford', 4);

-- works on

-- essn pno hours
INSERT INTO WORKS_ON
VALUES('123456789', 1, 32.5);

INSERT INTO WORKS_ON
VALUES('123456789', 2, 7.5);

INSERT INTO WORKS_ON
VALUES('666884444', 3, 40);

INSERT INTO WORKS_ON
VALUES('453453453', 1, 20);

INSERT INTO WORKS_ON
VALUES('453453453', 2, 20);

INSERT INTO WORKS_ON
VALUES('333445555', 2, 10);

INSERT INTO WORKS_ON
VALUES('333445555', 3, 10);

INSERT INTO WORKS_ON
VALUES('333445555', 10, 10);

INSERT INTO WORKS_ON
VALUES('333445555', 20, 10);

INSERT INTO WORKS_ON
VALUES('999887777', 30, 30);

INSERT INTO WORKS_ON
VALUES('999887777', 10, 10);

INSERT INTO WORKS_ON
VALUES('987987987', 30, 5);

INSERT INTO WORKS_ON
VALUES('987654321', 30, 20);

INSERT INTO WORKS_ON
VALUES('987654321', 20, 15);

