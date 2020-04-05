import React from 'react';
import {Checkbox, Input, Form, Layout, Button, Modal, Icon, notification, Card, Spin, Tooltip } from "antd";

export default class RoomCreator extends React.Component {
    constructor(props) {
        super(props);

        this.state = this.initState();
    }

    initState() {
        return { meetingCreated: false, roomName: "", ssn: "", ssns: [], buttonEnabled: false };
    }

    handleChange = (event) => {
        const ssn = event.target.value;
        let buttonEnabled = /^\d{12}$/.test(ssn);
        this.setState({ ssn, buttonEnabled });
    }

    addSsn = (event) => {
        let ssns = this.state.ssns;
        ssns.push(this.state.ssn);

        this.setState({ ssn: '', ssns, buttonEnabled: false });
    }

    handleChangeMeeting = (event) => {
        this.setState({ roomName: event.target.value });
    }

    addMeeting =  async () => {
        const { roomName, ssns } = this.state;
        let requestOptions = {
            method: 'POST',
            headers: { 'content-type': 'application/json' },
            body: JSON.stringify({ RoomName : roomName, PersonNumbers: ssns})
        };

        var res = await fetch('meeting', requestOptions);

        

        var meetingCreated = res.status == 200;

        this.setState({ meetingCreated });
    };

    clearState = () => {
        this.setState(this.initState());
    }

    render() {

        return (					<Card title="Welcome to HoldSpace meeting creation" className="app-login-card box-shadow" headStyle={{backgroundColor: 'rgba(255, 255, 255, 0.7)', borderBottom: '1px solid #BBB'}}>

            {!this.state.meetingCreated && (
					<div style={{position:'relative', height:'100%'}}>


               <h1> Meeting name:</h1> <br/>
					<input class="input" value={this.state.roomName} onChange={this.handleChangeMeeting} /><br />
					<br/>
                   <h1>SSN:</h1> <br/><input class="input" placeholder="YYYYMMDDXXXX"value={this.state.ssn} onChange={this.handleChange} type="text" />

					<button class="button-add-ssn" disabled={!this.state.buttonEnabled} onClick={this.addSsn}>Add SSN</button>
                    <ul style={{paddingTop:'40px'}} >
                    {this.state.ssns.map(ssn => (<li style={{height:'40px'}} key={ssn}>{ssn}<button class='button-add-ssn' >Remove</button></li>))}
									<br/>

                    </ul>
					<br/>
					<br/>

                    <button class="room-button"  onClick={this.addMeeting} disabled={!(this.state.ssns.length > 0 && this.state.roomName.length > 0)}>Create meeting</button>
					<br/>

				</div>)}
            {this.state.meetingCreated && (
					<div>Your meeting room was created. <button  onClick={this.clearState}>Create a new meeting!</button></div>
            )}
									</Card>);				
    };
}
