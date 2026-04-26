import XCTest
@testable import copyagent

final class CopyagentDaemonClientTests: XCTestCase {
  func testMissingExecutableReturnsError() {
    let client = CopyagentDaemonClient(executablePath: "/tmp/copyagentd-does-not-exist")

    XCTAssertThrowsError(try client.serviceStatus()) { error in
      XCTAssertEqual(error as? CopyagentDaemonClientError, .executableNotFound("/tmp/copyagentd-does-not-exist"))
    }
  }
}
