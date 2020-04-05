import React, { Component } from 'react';
import { Layout } from './components/Layout';
import RoomCreator from './components/RoomCreator'
import './custom.css'

export default class App extends Component {
  static displayName = App.name;

  render () {
    return (
      <Layout>
        <RoomCreator>
        </RoomCreator>
      </Layout>
    );
  }
}
