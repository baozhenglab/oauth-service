#!/usr/bin/env groovy

pipeline {
    agent any
    tools {
        go 'Go'
    }
    environment {
        GO111MODULE = 'on'
        GITHUB_USER_AND_TOKEN = credentials('GITHUB_USER_TOKEN')
        APP_NAME = 'oauth-dating'
        DEPLOY_PATH = '/var/lib/jenkins/deploy'
        EC2_IP = '172.31.25.127'
        PEM_FILE = 'dating-server.pem'
        myImg = ''
    }
    stages {
        stage('Mod Download') {
            steps {
                sh 'git config --global url."https://${GITHUB_USER_AND_TOKEN}@github.com/".insteadOf "https://github.com/"'
                sh 'go mod download'
            }
        }

        stage('Compile') {
            steps {
                sh 'CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .'
            }
        }

        stage('Docker build') {
            steps {
                script {
                    myImg = docker.build('${APP_NAME}', '-f ./deploy/Dockerfile .')
                    sh "docker save -o ${DEPLOY_PATH}/${APP_NAME}.tar ${APP_NAME}"
                }
            }
        }

        stage('Deploy') {
            steps {
                script {
                    if (env.GIT_BRANCH == "origin/staging") {
                        sh "scp -o StrictHostKeyChecking=no -i ${DEPLOY_PATH}/${PEM_FILE} ${DEPLOY_PATH}/${APP_NAME}.tar ec2-user@${EC2_IP}:/home/ec2-user/dating/stg"
                    }

                    if (env.GIT_BRANCH == "origin/master") {
                        sh "scp -o StrictHostKeyChecking=no -i ${DEPLOY_PATH}/${PEM_FILE} ${DEPLOY_PATH}/${APP_NAME}.tar ec2-user@${EC2_IP}:/home/ec2-user/dating/prd"
                    }
                }
            }
        }

        stage('Run') {
            steps {
                script {
                    if (env.GIT_BRANCH == "origin/staging") {
                        sh "ssh -o StrictHostKeyChecking=no -i ${DEPLOY_PATH}/${PEM_FILE} ec2-user@${EC2_IP} 'bash -s' < ./deploy/staging/run.sh"
                    }

                    if (env.GIT_BRANCH == "origin/master") {
                        sh "ssh -o StrictHostKeyChecking=no -i ${DEPLOY_PATH}/${PEM_FILE} ec2-user@${EC2_IP} 'bash -s' < ./deploy/master/run.sh"
                    }
                }
            }
        }
    }
}
