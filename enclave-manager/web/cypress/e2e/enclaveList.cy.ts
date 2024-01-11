describe("Enclave List", () => {
  let enclaveName: string = "unknown";

  beforeEach(() => {
    enclaveName = `cyrpress-test-${Math.floor(Math.random()*100000)}`
  })

  afterEach(() => {
    cy.deleteEnclave(enclaveName)
  })

  it("Can create and update an enclave", () => {
    cy.goToEnclaveList()
    cy.contains("Enclaves");

    // Create a Postgres enclave
    cy.contains("New Enclave").click();
    cy.focused().type("postgres");
    const configurationDrawer = cy.contains("[role='dialog']", "Enclave Configuration")
    configurationDrawer.contains("Postgres Package").click()

    configurationDrawer.focusInputWithLabel("Enclave name")
      .type(enclaveName)

    cy.contains("button", "Run").click();

    cy.url({timeout: 10 * 1000}).should("match", /enclave\/[^/]+\/logs/);

    cy.contains("button", "Edit").should("be.disabled");
    cy.contains("Validating", {timeout: 10 * 1000})
    cy.contains("Script completed", {timeout: 10 * 1000});
    cy.contains("button", "Edit").should("be.enabled");

    // Go to the enclave overview
    cy.contains("Go to Enclave Overview").click()
    cy.contains("[role='dialog'] button", "Continue").click();
    cy.url().should("match", /enclave\/[^/]+/);
    cy.findCardWithName("Name").contains(enclaveName)
    cy.contains("table", "postgres")
    cy.get("table").contains("button", "postgres").click();

    cy.url().should("match", /enclave\/[^/]+\/service\/[^/]+/);
    cy.findCardWithName("Name").contains("postgres")
    cy.findCardWithName("Status").contains("Running", {matchCase: false});

    // Update the postgres instance
    cy.contains("button", "Edit").click();
    cy.focusInputWithLabel("Max CPU").type("1024");
    cy.contains("button", "Update").click();
    cy.contains("Script completed", {timeout: 10 * 1000});
  });
});
