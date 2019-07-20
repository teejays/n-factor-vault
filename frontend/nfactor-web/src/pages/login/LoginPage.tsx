import React from 'react';

import {Typography, Row} from 'antd';

import {LoginForm} from '../../components/LoginForm';

export const LoginPage = () => {
  const handler = () => {};
  const {Title} = Typography;
  return (
    <>
      <Row
        type="flex"
        justify="center"
        align="middle"
        style={{textAlign: 'center', minHeight: 200}}>
        <Title>N-Factor Vault</Title>
      </Row>
      <Row
        type="flex"
        justify="center"
        align="middle"
        style={{textAlign: 'center', minHeight: 400}}>
        <LoginForm handleSuccess={() => {}} styles={{maxWidth: '300'}} />
      </Row>
    </>
  );
};

// Web Service call
interface LoginResponse {
  jwt: string;
}
