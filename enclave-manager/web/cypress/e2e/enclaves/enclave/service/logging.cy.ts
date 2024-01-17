describe("Service Logging", () => {
  const enclaveName = `cyrpress-test-${Math.floor(Math.random() * 100000)}`;
  before(() => {
    cy.createAndGoToEnclave(enclaveName);
  });
  after(() => {
    cy.deleteEnclave(enclaveName);
  });

  it("Should buffer logs", () => {
    cy.contains("button", "postgres").click();
    cy.contains("Logs", { matchCase: false }).click();

    // TODO: Implement this test once the WSS logging is used - cypress blocks the streamed
    // response until the request is closed.
  });
});
