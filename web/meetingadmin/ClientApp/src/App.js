import React, { Component } from 'react';
//import { Layout } from './components/Layout';
import RoomCreator from './components/RoomCreator'
import {Checkbox, Input, Form, Layout, Button, Modal, Icon, notification, Card, Spin, Tooltip } from "antd";

import './custom.css'
import './style.scss'
const { Header, Content, Footer, Sider } = Layout;

export default class App extends Component {
  static displayName = App.name;

  render () {
    return (

			<Layout className="app-layout">
			<Header className="app-header">
			<div className="app-header-left" >
            {/*<a href="https://pion.ly" target="_blank">*/}
              {/* <img src="/pion-logo.svg" className="app-logo-img" />*/}
			  <img src="/hsLogo.png" className="app-logo-img" />

              {/*</a>*/}
          </div>
		</Header>
        <Content className="app-center-layout box-shadow-inset">
			<RoomCreator>
			</RoomCreator>
			</Content>
			<Footer className="app-footer">
			</Footer>
		</Layout>
			

    );
  }
}
