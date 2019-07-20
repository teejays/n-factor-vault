import React from 'react';

import {Typography, Row} from 'antd';

import {SignupForm} from '../../components/SignupForm';

export const SignupPage = () => {
  const handler = () => {};
  const {Title} = Typography;
  return (
    <>
      <Row
        type="flex"
        justify="center"
        align="middle"
        style={{textAlign: 'center', minHeight: 200}}>
        <Title>N-Factor Vault - Signup</Title>
      </Row>
      <Row
        type="flex"
        justify="center"
        align="middle"
        style={{textAlign: 'center', minHeight: 400}}>
        <SignupForm
          handleSuccess={() => {
            console.log('Signup Up - Try logging in');
          }}
          styles={{maxWidth: '300'}}
        />
      </Row>
    </>
  );
};
