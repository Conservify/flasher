timestamps {
    node () {
        stage ('git') {
            checkout([$class: 'GitSCM', branches: [[name: '*/master']], userRemoteConfigs: [[url: 'https://github.com/Conservify/flasher.git']]])
        }

        stage ('build') {
            sh """
go get -u go.bug.st/serial.v1
go get -u github.com/Conservify/tooling
make clean
make
cp build/linux-amd64/flasher ~/workspace/bin
"""
        }
    }
}
