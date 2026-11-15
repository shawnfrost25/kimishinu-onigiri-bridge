# ==========================================
# SECURITY GROUPS
# ==========================================

resource "aws_security_group" "backend_sg" {
  name        = "kimishinu-${var.env}-sg-backend"
  description = "What is allowed to enter and leave our backend"
}

resource "aws_security_group" "database_sg" {
  name        = "kimishinu-${var.env}-sg-database"
  description = "What is allowed to enter and leave our database"
}

# ==========================================
# INGRESS / EGRESS RULES
# ==========================================

resource "aws_vpc_security_group_egress_rule" "backend_to_discord" {
  security_group_id = aws_security_group.backend_sg.id
  ip_protocol       = "tcp"
  description       = "Allow backend to send https requests"

  from_port = 443
  to_port   = 443
  cidr_ipv4 = "0.0.0.0/0"
}

resource "aws_vpc_security_group_egress_rule" "backend_to_database" {
  security_group_id = aws_security_group.backend_sg.id
  ip_protocol       = "tcp"
  description       = "Allow backend to interact with postgresql"

  from_port                    = 5432
  to_port                      = 5432
  referenced_security_group_id = aws_security_group.database_sg.id
}

resource "aws_vpc_security_group_ingress_rule" "ssh_to_backend" {
  security_group_id = aws_security_group.backend_sg.id
  ip_protocol       = "tcp"
  description       = "allow ssh"

  from_port = 22
  to_port   = 22
  cidr_ipv4 = "0.0.0.0/0"
}

resource "aws_vpc_security_group_ingress_rule" "database_from_backend" {
  security_group_id = aws_security_group.database_sg.id
  ip_protocol       = "tcp"
  description       = "Expose database for backend interaction"

  from_port                    = 5432
  to_port                      = 5432
  referenced_security_group_id = aws_security_group.backend_sg.id
}
