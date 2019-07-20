import React from 'react';

import {Typography, Row} from 'antd';

import {LoginForm} from '../../components/LoginForm';

export const LoginPage = () => {
  const {Title} = Typography;
  return (
    <>
      <Row
        type="flex"
        justify="center"
        align="middle"
        style={{textAlign: 'center', minHeight: 600}}>
        <LoginForm handleLogin={() => {}} maxWidth={300} />
      </Row>
    </>
  );
};
