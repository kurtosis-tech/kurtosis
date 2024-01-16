/// <reference types="cypress" />

export {};

Cypress.Commands.add("goToEnclaveList", () => {
  return cy.visit("http://localhost:4000");
});

Cypress.Commands.add("goToCatalog", () => {
  cy.visit("http://localhost:9711").get('nav button[aria-label="View catalog"]').click();
});

Cypress.Commands.add("createAndGoToEnclave", (enclaveName: string) => {
  cy.goToEnclaveList();
  cy.contains("Enclaves");

  // Create a Postgres enclave
  cy.contains("New Enclave").click();
  cy.focused().type("postgres");
  const configurationDrawer = cy.contains("[role='dialog']", "Enclave Configuration");
  configurationDrawer.contains("Postgres Package").click();

  configurationDrawer.focusInputWithLabel("Enclave name").type(enclaveName);

  cy.contains("button", "Run").click();

  cy.url({ timeout: 10 * 1000 }).should("match", /enclave\/[^/]+\/logs/);

  cy.contains("button", "Edit").should("be.disabled");
  cy.contains("Validating", { timeout: 10 * 1000 });
  cy.contains("Script completed", { timeout: 10 * 1000 });
  cy.contains("button", "Edit").should("be.enabled");

  // Go to the enclave overview
  cy.contains("Go to Enclave Overview").click();
  cy.contains("[role='dialog'] button", "Continue").click();
  cy.url().should("match", /enclave\/[^/]+/);
});

Cypress.Commands.add("focusInputWithLabel", (label: string) => {
  cy.contains("label", label, { matchCase: false }).parents(".chakra-form-control").first().find("input").click();
});

Cypress.Commands.add("findCardWithName", (name: string) => {
  return cy.contains(".chakra-card", name, { matchCase: false });
});

Cypress.Commands.add("deleteEnclave", (enclaveName: string) => {
  cy.goToEnclaveList();
  cy.contains("tr", enclaveName).find(".chakra-checkbox").click();
  cy.contains("button", "Delete").click();
  cy.contains("[role='dialog'] button", "Delete").click();
  cy.contains("[role='dialog']", "Delete", { timeout: 10 * 1000 }).should("not.exist");
  cy.contains("tr", enclaveName).should("not.exist");
});

declare global {
  namespace Cypress {
    interface Chainable {
      createAndGoToEnclave(enclaveName: string): Chainable<void>;
      deleteEnclave(enclaveName: string): Chainable<void>;
      focusInputWithLabel(label: string): Chainable<void>;
      findCardWithName(label: string): Chainable<JQuery<HTMLElement>>;
      goToCatalog(): Chainable<void>;
      goToEnclaveList(): Chainable<AUTWindow>;
    }
  }
}
