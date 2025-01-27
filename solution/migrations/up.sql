CREATE TABLE if not exists companies
(
    company_id uuid NOT NULL,
    company_name character varying(50) NOT NULL,
    email character varying(120) NOT NULL,
    password bytea NOT NULL,
    PRIMARY KEY (company_id, company_name, email)
);