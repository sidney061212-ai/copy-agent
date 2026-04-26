import Cocoa

class About {
  private let productCredits = NSAttributedString(
    string: "Lightweight clipboard and remote workflow agent for the AI era.",
    attributes: [NSAttributedString.Key.foregroundColor: NSColor.labelColor]
  )

  private let workflowCredits = NSAttributedString(
    string: "Move text, images, files, and tasks between phone and computer.",
    attributes: [NSAttributedString.Key.foregroundColor: NSColor.labelColor]
  )

  private let attributionCredits = NSAttributedString(
    string: "Third-party attribution details are available in NOTICE.md and LICENSE.",
    attributes: [NSAttributedString.Key.foregroundColor: NSColor.secondaryLabelColor]
  )

  private var links: NSMutableAttributedString {
    let string = NSMutableAttributedString(string: "GitHub | Documentation",
                                           attributes: [NSAttributedString.Key.foregroundColor: NSColor.labelColor])
    string.addAttribute(.link, value: "https://github.com/yu/copyagent", range: NSRange(location: 0, length: 6))
    string.addAttribute(.link, value: "https://github.com/yu/copyagent/tree/main/docs", range: NSRange(location: 9, length: 13))
    return string
  }

  private var credits: NSMutableAttributedString {
    let credits = NSMutableAttributedString(string: "",
                                            attributes: [NSAttributedString.Key.foregroundColor: NSColor.labelColor])
    credits.append(links)
    credits.append(NSAttributedString(string: "\n\n"))
    credits.append(productCredits)
    credits.append(NSAttributedString(string: "\n"))
    credits.append(workflowCredits)
    credits.append(NSAttributedString(string: "\n\n"))
    credits.append(attributionCredits)
    credits.setAlignment(.center, range: NSRange(location: 0, length: credits.length))
    return credits
  }

  @objc
  func openAbout(_ sender: NSMenuItem?) {
    NSApp.activate(ignoringOtherApps: true)
    NSApp.orderFrontStandardAboutPanel(options: [NSApplication.AboutPanelOptionKey.credits: credits])
  }
}
