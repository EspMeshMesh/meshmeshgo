import fakeDataProvider from "ra-data-fakerest";
import simpleRestProvider from 'ra-data-simple-rest';
import { DataProvider, withLifecycleCallbacks } from "react-admin";

export const myFakeDataProvider = fakeDataProvider({
    nodes: [
        { id: 0, tag: 'Hello, world!' },
        { id: 1, tag: 'FooBar' },
    ],
    links: [
        { id: 0, post_id: 0, tag: 'John Doe', body: 'Sensational!' },
        { id: 1, post_id: 0, tag: 'Jane Doe', body: 'I agree' },
    ],
})

export const dataProvider = withLifecycleCallbacks(simpleRestProvider('/api/v1'), [
    {
        resource: 'nodes',
        /*beforeUpdate: async (params: any, _: DataProvider) => {
            let base64firmware = null;
            const newFirmware = params.data.firmware;
            if (newFirmware.rawFile instanceof File) {
                base64firmware = await convertFileToBase64(newFirmware);
            } else {
            }

            return {
                ...params,
                data: {
                    ...params.data,
                    firmware: base64firmware,
                },
            };
        },*/
    }
]);

/**
 * Convert a `File` object returned by the upload input into a base 64 string.
 * That's not the most optimized way to store images in production, but it's
 * enough to illustrate the idea of dataprovider decoration.
 */
/*const convertFileToBase64 = (file: any) =>
    new Promise((resolve, reject) => {
        const reader = new FileReader();
        reader.onload = () => resolve(reader.result);
        reader.onerror = reject;
        reader.readAsDataURL(file.rawFile);
});*/