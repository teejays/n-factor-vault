import * as React from 'react';
import {Link} from 'react-router-dom';

import {Form, Icon, Input, Button, Row} from 'antd';
import {WrappedFormUtils} from 'antd/lib/form/Form';

import {makeRequest} from '../utils/api';

export interface FormProps {
  form: WrappedFormUtils;
}

interface Props extends FormProps {
  // handleSuccess is what handles the successful login request
  handleSuccess?: Function;
  styles?: React.CSSProperties;
}

// LoginForm renders the actual form needed for login. It takes a handlerLogin() prop, and no stores no state.
// The first param in 'React.Component<Props, {}>', 'Props' are the props and the second param '{}' is the type of the state.
class LoginFormBase extends React.Component<Props, {}> {
  handleSubmit = (e: React.FormEvent) => {
    // I don't know what this does
    e.preventDefault();
    // Validate and log the fields
    this.props.form.validateFields(async (err, values) => {
      if (err) {
        console.error('error validating form: ', err);
        return;
      }

      console.log('Received values of form: ', values);

      // HTTP Post request
      const response = await makeRequest({
        path: 'login',
        method: 'POST',
        body: values,
      });

      console.log('Login response: ', response);

      // Invoke other handler
      if (this.props.handleSuccess) {
        this.props.handleSuccess();
      }
    });
  };

  render() {
    const {getFieldDecorator} = this.props.form;
    return (
      <Row type="flex" justify="center" align="middle">
        <Form
          className="login-form"
          onSubmit={this.handleSubmit}
          style={this.props.styles ? this.props.styles : {}}>
          <Form.Item>
            {getFieldDecorator('email', {
              rules: [{required: true, message: 'Please input your username!'}],
            })(
              <Input
                prefix={<Icon type="mail" style={{color: 'rgba(0,0,0,.25)'}} />}
                placeholder="Email"
              />,
            )}
          </Form.Item>
          <Form.Item>
            {getFieldDecorator('password', {
              rules: [{required: true, message: 'Please input your Password!'}],
            })(
              <Input
                prefix={<Icon type="lock" style={{color: 'rgba(0,0,0,.25)'}} />}
                type="password"
                placeholder="Password"
              />,
            )}
          </Form.Item>
          <Form.Item>
            <Button
              type="primary"
              htmlType="submit"
              className="login-form-button"
              style={{width: '100%'}}>
              Log in
            </Button>
            Or <Link to="/signup">register now! (coming soon)</Link>
          </Form.Item>
        </Form>
      </Row>
    );
  }
}

export const LoginForm = Form.create<Props>()(LoginFormBase);
