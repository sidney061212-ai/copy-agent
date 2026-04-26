import AppKit
import Defaults

enum MenuIcon: String, CaseIterable, Identifiable, Defaults.Serializable {
  case copyagent

  var id: Self { self }

  var image: NSImage {
    let image = switch self {
    case .copyagent:
      NSImage(named: .copyagentStatusBar)!
    }
    image.size = NSSize(width: 18, height: 18)
    return image
  }
}
