@Library('conservify') _

conservifyProperties()

timestamps {
    node () {
        stage ('git') {
            checkout scm
        }

        stage ('build') {
            withEnv(["PATH+GOLANG=${tool 'golang-amd64'}/bin"]) {
                sh """
make clean
make
cp build/linux-amd64/flasher ~/workspace/bin
"""
            }
        }

        stage ("archive") {
            dir ("build") {
                archiveArtifacts "*.zip"
            }
        }
    }
}
