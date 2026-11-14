data "aws_ssm_parameter" "db_password" {
  name            = "/kimishinu/prod/db-password"
  with_decryption = true
}

resource "aws_db_instance" "mimi_database_engine" {
  allocated_storage     = 20
  max_allocated_storage = 100
  engine                = "postgres"
  engine_version        = "18"
  instance_class        = "db.t4g.micro"
  username              = "kimishinu_db"
  password              = data.aws_ssm_parameter.db_password.value

  vpc_security_group_ids = [aws_security_group.database_sg.id]

  skip_final_snapshot = true
}
