describe("Catalog", () => {
  it("Lists the catalog", () => {
    cy.goToCatalog();
    cy.contains("Package Catalog");
    cy.contains("Most Ran");
    cy.contains("Most starred");
    cy.contains("All");
    cy.contains("Postgres");
  });

  it("Searches with ctrl+f", () => {
    cy.goToCatalog();
    cy.contains("Package Catalog");
    cy.get("body").type("{ctrl}f");
    cy.focused().type("postgres");
    cy.contains("1 Matches");
  });

  it("Opens packages details", () => {
    cy.goToCatalog();
    cy.contains("Postgres").click();
    cy.url().should("match", /catalog\/.*postgres/);
    cy.contains("button", "Run").click();
    cy.contains("[role='dialog']", "Postgres Package").contains("Run");
  });

  it("Can save packages", () => {
    cy.goToCatalog();
    // Trigger click rather than use `click` so that it is exactly the button that is clicked
    cy.contains("a", "Postgres").find("button[aria-label*='save']").trigger("click");

    // Now, remove it from favourites
    cy.contains(".chakra-card", "Saved").contains("a", "Postgres").find("button[aria-label*='save']").trigger("click");
    cy.contains(".chakra-card", "Saved").should("not.exist");
  });
});
