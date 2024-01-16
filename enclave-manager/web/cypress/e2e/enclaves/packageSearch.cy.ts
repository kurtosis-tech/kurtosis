describe("New Enclave package search", () => {
  it("Can find an enclave by exact name", () => {
    cy.goToEnclaveList();
    cy.contains("New Enclave").click();
    cy.contains("Package Template Repo").should("not.exist");
    cy.focused().type("github.com/kurtosis-tech/package-template-repo");
    cy.contains("Package Template Repo");
  });
});
