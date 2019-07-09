import * as React from 'react';

import {Form, Icon, Input, Button, Row} from 'antd';
import { WrappedFormUtils,  } from 'antd/lib/form/Form'

import createReactContext from '@ant-design/create-react-context';

export interface FormProps {
    form: WrappedFormUtils
}

interface  Props extends FormProps {
    // handlerLogin is what handles the login request
    handleLogin(): void;
    maxWidth?: number;
}

// LoginForm renders the actual form needed for login. It takes a handlerLogin() prop, and no stores no state.
// The first param in 'React.Component<Props, {}>', 'Props' are the props and the second param '{}' is the type of the state.
class LoginFormBase extends React.Component<Props, {}> {
    
    handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        this.props.form.validateFields((err, values) => {
            if (!err) {
                console.log('Received values of form: ', values);
            }
        });
        this.props.handleLogin();
      };

    render() {
        const { getFieldDecorator } = this.props.form;
        return (
            <Row type="flex" justify="center" align="middle">
                
                <Form className="login-form" onSubmit={this.handleSubmit} style={{
                    maxWidth: this.props.maxWidth
                }}>
                    <Form.Item>
                        {getFieldDecorator('username', {
                            rules: [{ required: true, message: 'Please input your username!' }],
                        })(
                            <Input
                            prefix={<Icon type="user" style={{ color: 'rgba(0,0,0,.25)' }} />}
                            placeholder="Username"
                            />,
                        )}
                    </Form.Item>
                    <Form.Item>
                        {getFieldDecorator('password', {
                            rules: [{ required: true, message: 'Please input your Password!' }],
                        })(
                            <Input
                            prefix={<Icon type="lock" style={{ color: 'rgba(0,0,0,.25)' }} />}
                            type="password"
                            placeholder="Password"
                            />,
                        )}
                    </Form.Item>
                    <Form.Item>
                        <Button type="primary" htmlType="submit" className="login-form-button" style={{width: "100%"}}>
                            Log in
                        </Button>
                        Or <a href="">register now! (coming soon)</a>
                        </Form.Item>
                </Form>
            </Row>
        );
    }
    
}

export const LoginForm = Form.create<Props>()(LoginFormBase);