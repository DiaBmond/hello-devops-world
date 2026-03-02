# ==========================================
# 1. Terraform & Provider Setup
# ==========================================
terraform {
  required_version = ">= 1.5.0"

  backend "s3" {
    bucket = "hello-devops-tfstate-mon-998877"
    key    = "prod/terraform.tfstate"
    region = "ap-southeast-1"
  }

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.region
}

# ==========================================
# 2. Variables
# ==========================================
variable "region" {
  default = "ap-southeast-1"
}

variable "my_ip" {
  description = "Your public IP for SSH access"
  type        = list(string)
}

variable "instance_type" {
  default = "t3.micro"
}

variable "db_password" {
  description = "Password for the PostgreSQL database"
  type        = string
  sensitive   = true
}

# ==========================================
# 3. Ubuntu 22.04 AMI
# ==========================================
data "aws_ami" "ubuntu" {
  most_recent = true
  owners      = ["099720109477"]

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"]
  }
}

# ==========================================
# 4. Security Group
# ==========================================
resource "aws_security_group" "web_sg" {
  name        = "hello-devops-sg"
  description = "Allow SSH (restricted) and HTTP traffic"

  # SSH restricted to your IP
  ingress {
    from_port = 22
    to_port   = 22
    protocol  = "tcp"
    # cidr_blocks = ["${var.my_ip}/32"]
    cidr_blocks = [
      for ip in var.my_ip : "${ip}/32"
    ]
  }

  # Public HTTP
  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Egress (allow all outbound)
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "hello-devops-sg"
    Env  = "dev"
  }
}

# ==========================================
# 5. EC2 Instance 
# ==========================================
resource "aws_instance" "app_server" {
  ami                    = data.aws_ami.ubuntu.id
  instance_type          = var.instance_type
  vpc_security_group_ids = [aws_security_group.web_sg.id]

  user_data = <<-EOF
              #!/bin/bash
              apt update -y
              apt install -y docker.io
              systemctl enable docker
              systemctl start docker

              docker network create devops-net

              docker run -d \
                --name database \
                --network devops-net \
                -e POSTGRES_USER=postgres \
                -e POSTGRES_PASSWORD=${var.db_password} \
                -e POSTGRES_DB=myapp \
                postgres:15-alpine

              docker run -d \
                --name api \
                --network devops-net \
                -p 80:8080 \
                -e DB_DSN="postgres://postgres:${var.db_password}@database:5432/myapp?sslmode=disable" \
                ghcr.io/diabmond/hello-devops-app:latest
              EOF

  tags = {
    Name = "hello-devops-server"
    Env  = "dev"
  }
}

output "server_public_ip" {
  description = "Public IP of the EC2 instance"
  value       = aws_instance.app_server.public_ip
}
