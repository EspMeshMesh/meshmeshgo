import simpleRestProvider from 'ra-data-simple-rest';
import { DataProvider } from 'react-admin';

const baseDataProvider =  simpleRestProvider(window.location.pathname + 'api/v1');

export const dataProvider = { 
    ...baseDataProvider,
    hello: () => {
        return fetch(window.location.pathname + 'api/v1/hello', {
            method: 'GET',
        }).then(response => response.json());
    }
};

export interface MyDataProvider extends DataProvider {
    hello: () => Promise<Record<string, any>>;
}