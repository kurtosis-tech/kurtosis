import {createContext, Dispatch, SetStateAction, useContext, useState} from 'react';

const defaultState = {
    userId: undefined,
    externalUserId: undefined,
    apiKey: undefined,
    instanceId: undefined,
    instanceStatus: undefined,
    jwtToken: undefined,
    apiProtocol: undefined,
    apiHost: undefined,
    apiPort: undefined,
    webUrl: undefined,
}

export const AppContext = createContext({
    appData: defaultState,
    setAppData: () => {
    },
});

export const AppStateProvider = (props) => {
    const [appData, setAppDataTemp] = useState(defaultState);

    const setAppData = (appContext) => {
        const combined = {...appData, ...appContext}
        setAppDataTemp(combined);
    };
    const value = {appData, setAppData}
    return <AppContext.Provider value={value}>{props.children}</AppContext.Provider>;
};

export const useAppContext = () => {
    const appContext = useContext(AppContext);
    if (appContext === undefined) {
        throw new Error('AppContext must be inside a AppStateProvider');
    }
    return appContext;
};

export default AppContext;
