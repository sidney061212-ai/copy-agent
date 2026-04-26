import Observation

@Observable
class SoftwareUpdater {
  var automaticallyChecksForUpdates = false

  func checkForUpdates() {
    // copyagent-ui-mac disables Sparkle until the release/signing pipeline exists.
  }
}
