import React, { Component } from 'react';
import { Container, Navbar, NavbarBrand } from 'reactstrap';
import './NavMenu.css';

export class NavMenu extends Component {
  static displayName = NavMenu.name;

  constructor (props) {
    super(props);
  }

  render () {
    return (
      <header>
        <Navbar className="navbar-expand-sm navbar-toggleable-sm ng-white border-bottom box-shadow mb-3 header" light>
          <Container>
                      <div className="app-header-left" >
            <a href="https://pion.ly" target="_blank">
              {/* <img src="/pion-logo.svg" className="app-logo-img" />*/}
			  <img src='./hsLogo.png' className="app-logo-img" />

            </a>
          </div>
          </Container>
        </Navbar>
      </header>
    );
  }
}
