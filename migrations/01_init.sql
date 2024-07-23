-- +goose Up

CREATE TABLE agents (
  id varchar(255) NOT NULL,
  updated DATETIME,
  active BOOLEAN,
  PRIMARY KEY(id)
);

CREATE TABLE operators (
  pipeline_id varchar(255) NOT NULL,
  operator_id  varchar(255) NOT NULL,
  state varchar(255) NOT NULL,
  container_id varchar(255),
  error varchar(255),
  agent_id varchar(255) NOT NULL,

  PRIMARY KEY(pipeline_id, operator_id),
  FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE NO ACTION
);

-- +goose Down
DROP TABLE agents;
DROP TABLE operators;