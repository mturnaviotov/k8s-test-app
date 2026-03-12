// Uses Declarative syntax to run commands inside a container.
pipeline {
    agent {
        kubernetes {
            // containerTemplate now deprecated, use 'yaml' block instead
            yaml '''
apiVersion: v1
kind: Pod
spec:
  containers:
  - name: shell
    image: golang:1.25-alpine
    privileged: true
    command:
    - sleep
    args:
    - infinity
    volumeMounts:
    - mountPath: /var/run/docker.sock
      name: docker-sock
    # securityContext:
      # ubuntu runs as root by default, it is recommended or even mandatory in some environments (such as pod security admission "restricted") to run as a non-root user.
      # runAsUser: 501
  volumes:
  - name: docker-sock
    hostPath:
      path: /var/run/docker.sock
      type: Socket
'''
            // Can also wrap individual steps:
            // container('shell') {
            //     sh 'hostname'
            // }
            defaultContainer 'shell'
            retries 2
        }
    }
    stages {
        stage('Prepare') {
            steps {
                checkout scmGit(branches: [[name: '*']],
                    extensions: [], userRemoteConfigs:
                    [[url: 'https://github.com/mturnaviotov/go-k8s-test-app.git']])
                sh '''
                  apk add --no-cache docker git
                  git config --global --add safe.directory "*"
                '''
                withCredentials([usernamePassword(credentialsId: 'dockerhub', passwordVariable: 'regpass', usernameVariable: 'reguser')]) {
                    sh 'docker login -u $reguser -p $regpass'
                }
            }
        }
        stage('Run Builds') {
            parallel {
                stage('Backend') {
                    steps {
                        sh '''
                          cd backend
                          tag=`git rev-parse --short HEAD`
                          docker buildx build --no-cache --platform linux/amd64,linux/arm64/v8 --push -t foxmuldercp/go-todo-backend:${tag} .
                        '''
                    }
                    post {
                        always {
                            sh 'echo Backend build finished. tests can be pushed here'
                        }
                    }
                }
                stage('Frontend') {
                    steps {
                        sh 'cd frontend; docker buildx build --no-cache --platform linux/amd64,linux/arm64/v8 --push -t foxmuldercp/go-todo-frontend:latest .'
                    }
                    post {
                        always {
                            sh 'echo Frontend build finished. tests can be pushed here'
                        }
                    }
                }
            }
        }
    }
}
