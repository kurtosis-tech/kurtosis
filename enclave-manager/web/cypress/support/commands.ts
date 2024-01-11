/// <reference types="cypress" />

export {};

Cypress.Commands.add("goToEnclaveList", () => {
  return cy.visit("http://localhost:9711")
})

Cypress.Commands.add('focusInputWithLabel', (label: string) => {
  cy.contains("label", label, {matchCase: false})
    .parents(".chakra-form-control")
    .first()
    .find("input")
    .click()
})

Cypress.Commands.add('findCardWithName', (name: string) => {
  return cy.contains(".chakra-card", name, {matchCase: false})
})

Cypress.Commands.add('deleteEnclave', (enclaveName: string) => {
  cy.goToEnclaveList();
  cy.contains("tr", enclaveName).find(".chakra-checkbox").click()
  cy.contains("button", "Delete").click();
  cy.contains("[role='dialog'] button", "Delete").click();
  cy.contains("[role='dialog']", "Delete", {timeout: 10 * 1000}).should("not.exist")
  cy.contains("tr", enclaveName).should("not.exist")
})

declare global {
  namespace Cypress {
    interface Chainable {
      deleteEnclave(enclaveName: string): Chainable<void>
      focusInputWithLabel(label: string): Chainable<void>
      findCardWithName(label: string): Chainable<JQuery<HTMLElement>>
      goToEnclaveList(): Chainable<AUTWindow>
    }
  }
}
