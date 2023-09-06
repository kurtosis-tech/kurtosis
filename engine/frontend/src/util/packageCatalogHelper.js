export const checkValidUndefinedType = (data) => {
    try {
        const val = yaml.load(data)
    } catch (ex) {
        return false;
    }
    return true;
}

export const checkValidStringType = (data) => {
    if (data === "undefined" || data.length === 0) {
        return false
    }

    try {
        const val = JSON.parse(data)
        if (typeof val === "string") {
            return true
        }
        return false
    } catch (ex) {
        if (data.includes("\"") || data.includes("\'")) {
            return false
        }
        return true;
    }
}

export const checkValidIntType = (data) => {
    const isNumeric = (value) => {
        return /^-?\d+$/.test(value);
    }
    if (data === "undefined") {
        return false
    }
    try {
        return isNumeric(data)
    } catch(ex) {
        return false
    }
}

export const checkValidFloatType = (data) => {
    const  isValidFloat = (value) => {
        return (/^-?[\d]*(\.[\d]+)?$/g).test(value);
    }
    if (data === "undefined") {
        return false
    }
    return isValidFloat(data)
}

export const checkValidBooleanType = (data) => {
    return ["TRUE", "FALSE"].includes(data)
}
