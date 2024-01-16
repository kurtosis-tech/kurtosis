describe("Enclave List", () => {
  let enclaveName: string = "unknown";

  beforeEach(() => {
    enclaveName = `cyrpress-test-${Math.floor(Math.random() * 100000)}`;
  });

  afterEach(() => {
    cy.deleteEnclave(enclaveName);
  });

  it("Can create and update an enclave", () => {
    cy.createAndGoToEnclave(enclaveName);
    cy.findCardWithName("Name").contains(enclaveName);
    cy.contains("table", "postgres");
    cy.get("table").contains("button", "postgres").click();

    cy.url().should("match", /enclave\/[^/]+\/service\/[^/]+/);
    cy.findCardWithName("Name").contains("postgres");
    cy.findCardWithName("Status").contains("Running", { matchCase: false });

    // Update the postgres instance
    cy.contains("button", "Edit").click();
    cy.focusInputWithLabel("Max CPU").type("1024");
    cy.contains("button", "Update").click();
    cy.contains("Script completed", { timeout: 10 * 1000 });
  });

  it("Shows a new enclave in the list", () => {
    cy.goToEnclaveList();
    cy.contains("tr", enclaveName).should("not.exist");

    cy.createAndGoToEnclave(enclaveName);

    cy.goToEnclaveList();

    cy.contains("tr", enclaveName).should("exist");
    cy.contains("tr", enclaveName).contains("Running");
    cy.contains("tr", enclaveName).contains("Postgres Package");
  });
});
