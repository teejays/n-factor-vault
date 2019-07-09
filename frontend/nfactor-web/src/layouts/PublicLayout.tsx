import * as React from 'react';

import {Layout} from 'antd';

interface PublicLayoutProps {
    header: React.ReactNode;
    content: React.ReactNode;
    footer: React.ReactNode;
}

const { Header, Footer, Content } = Layout;

export const PublicLayout = (props: PublicLayoutProps) => {
    return (
        <Layout className="layout">
            <Header>{props.header}</Header>
            <Content>{props.content}</Content>
            <Footer>{props.footer}</Footer>
        </Layout>
    )
}


