provider "aws" {
  region = var.region
}

resource "aws_instance" "kimishinu-backend" {
  ami                    = data.aws_ami.ubuntu_ami
  instance_type          = var.backend_instance_type
  vpc_security_group_ids = [aws_security_group.backend_sg.id]

  root_block_device {
    volume_type           = "gp3"
    volume_size           = 10
    encrypted             = true
    delete_on_termination = true
  }

  tags = {
    Name = "kimishinu-poll-backend"
  }
}

output "vm_public_ip" {
  value = aws_instance.kimishinu-backend.public_ip
}
