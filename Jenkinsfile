@Library('conservify') _

conservifyProperties()

timestamps {
    node () {
        stage ('git') {
            checkout scm
        }

        stage ('build') {
            sh """
go get -u go.bug.st/serial.v1
go get -u github.com/conservify/tooling
make clean
make
cp build/linux-amd64/flasher ~/workspace/bin
"""
        }

        stage ("archive") {
            dir ("build") {
                archiveArtifacts "*.zip"
            }
        }
    }
}
