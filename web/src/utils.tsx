
export const formatNodeId = (id: any) => {
    if (typeof id === 'number') {
        return "N" + id.toString(16).toUpperCase().padStart(6, '0')
    } else if (typeof id === 'string') {
        if (id.startsWith('0x')) {
            return "N" + id.slice(2).toUpperCase().padStart(6, '0')
        }
    }
    return id
};

const _validateNodeId = (id: any) => {
    if (typeof id === 'string') {
        if (!id.startsWith('N')) {
            return false
        }
        if (id.slice(1).length !== 6) {
            return false
        }
        if (isNaN(parseInt(id.slice(1), 16))) {
            return false
        }
        return true
    }
    return false
};

export const validateNodeId = (values: any) => {
    if (!_validateNodeId(values.ID)) {
        return {'ID': 'Invalid node ID'}
    }
    return {}
}

export const parseNodeId = (id: any) => {
    if (typeof id === 'string') {
        if (id.startsWith('N')) {
            return parseInt(id.slice(1), 16)
        }
    }
    return id
};


