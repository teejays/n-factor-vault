// TODO: Make this file a module with functions (?)
import React from 'react';

// getURL provides the full URL for the API http request
const getURL = (path: string) => {
  const baseURL = 'localhost:8080/v1';
  return `http://${baseURL}/${path}`;
};
interface RequestProps {
  path: string;
  method: 'GET' | 'POST' | 'PUT' | 'DELETE';
  body?: any;
}

export const makeRequest = (props: RequestProps) =>
  (async () => {
    // Request Object
    const req = {
      method: 'POST',
      headers: {
        Accept: 'application/json',
        'Content-Type': 'application/json',
      },
      body: '',
    };
    if (props.body) {
      req.body = JSON.stringify(props.body);
    }
    // Get Response
    console.debug(`api: Making ${props.method} request to ${props.path}`);
    const rawResponse = await fetch(getURL(props.path), req);
    const content = await rawResponse.json();

    console.log(content);
    return content;
  })();
