timestamps {
    node () {
        stage ('git') {
            checkout([$class: 'GitSCM', branches: [[name: '*/master']], userRemoteConfigs: [[url: 'https://github.com/Conservify/flasher.git']]])
        }

        stage ('build') {
            sh """
go get go.bug.st/serial.v1
make clean
make
cp flasher ~/workspace/bin
"""
        }

        stage ('archive') {
            archiveArtifacts artifacts: 'flasher'
        }
    }
}
