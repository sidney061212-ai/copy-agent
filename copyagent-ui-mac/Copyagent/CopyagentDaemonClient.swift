import Foundation

struct CopyagentCommandResult: Equatable {
  let exitCode: Int32
  let output: String
}

enum CopyagentDaemonClientError: Error, Equatable {
  case executableNotFound(String)
}

final class CopyagentDaemonClient {
  private let executablePath: String

  init(executablePath: String = CopyagentDaemonClient.defaultExecutablePath()) {
    self.executablePath = executablePath
  }

  func doctor() throws -> CopyagentCommandResult {
    try run(["doctor"])
  }

  func serviceStatus() throws -> CopyagentCommandResult {
    try run(["service", "status"])
  }

  func serviceLogs() throws -> CopyagentCommandResult {
    try run(["service", "logs"])
  }

  func run(_ arguments: [String]) throws -> CopyagentCommandResult {
    guard FileManager.default.isExecutableFile(atPath: executablePath) else {
      throw CopyagentDaemonClientError.executableNotFound(executablePath)
    }

    let process = Process()
    process.executableURL = URL(fileURLWithPath: executablePath)
    process.arguments = arguments

    let outputPipe = Pipe()
    process.standardOutput = outputPipe
    process.standardError = outputPipe

    try process.run()
    process.waitUntilExit()

    let output = String(data: outputPipe.fileHandleForReading.readDataToEndOfFile(), encoding: .utf8) ?? ""
    return CopyagentCommandResult(exitCode: process.terminationStatus, output: output)
  }

  static func defaultExecutablePath() -> String {
    let homeDirectory = FileManager.default.homeDirectoryForCurrentUser.path
    for path in [
      "\(homeDirectory)/.local/bin/copyagentd",
      "/opt/homebrew/bin/copyagentd",
      "/usr/local/bin/copyagentd"
    ] where FileManager.default.isExecutableFile(atPath: path) {
      return path
    }
    return "/usr/local/bin/copyagentd"
  }
}
