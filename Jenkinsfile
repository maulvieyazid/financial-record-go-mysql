pipeline {
    agent { label 'dev' }

    stages {
        stage('Pull SCM') {
            steps {
                git branch: 'main', url: 'https://github.com/maulvieyazid/financial-record-go-mysql.git'
            }
        }
        
        stage('Build') {
            steps {
                sh'''
                cd app
                go mod tidy
                '''
            }
        }
        
        stage('Code Review') {
            steps {
                sh'''
                cd app
                sonar-scanner \
                    -Dsonar.projectKey=app-mlv \
                    -Dsonar.sources=. \
                    -Dsonar.host.url=http://172.23.10.12:9000 \
                    -Dsonar.token=sqp_f14d32262db72dfd105cd3951e42442b27d52bcf
                '''
            }
        }
        
        stage('Deploy') {
            steps {
                sh'''
                docker compose up --build -d
                '''
            }
        }
        
        stage('Backup') {
            steps {
                 sh 'docker compose push' 
            }
        }
    }
}