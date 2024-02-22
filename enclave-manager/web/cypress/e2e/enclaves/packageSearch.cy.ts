describe("New Enclave package search", () => {
  it("Can find an enclave by exact name", () => {
    cy.goToEnclaveList();
    cy.contains("New Enclave").click();
    cy.contains("Package Template Repo").should("not.exist");
    cy.focused().type("github.com/kurtosis-tech/package-template-repo");
    cy.contains("Package Template Repo");
  });

  it("Can find an enclave with a github url", () => {
    cy.goToEnclaveList();
    cy.contains("New Enclave").click().wait(500);
    cy.contains("Exact Match").should("not.exist");
    cy.focused().type("https://github.com/kurtosis-tech/awesome-kurtosis/tree/main/redis-voting-app");
    cy.contains("Exact Match");
    cy.contains("Redis Voting App");
  });
});
