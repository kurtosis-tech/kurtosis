describe("Enclave List", () => {
  it("passes", () => {
    cy.visit("http://localhost:9711");
    cy.contains("Enclaves");
    cy.contains("New Enclave").click();
  });
});
