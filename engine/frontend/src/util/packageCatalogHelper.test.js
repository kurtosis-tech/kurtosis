import { checkValidFloatType, checkValidStringType, checkValidBooleanType, checkValidUndefinedType, checkValidIntType} from './packageCatalogHelper';


it("test whether boolean is valid or not", () => {
    expect(checkValidBooleanType("TRUE")).toEqual(true)
})