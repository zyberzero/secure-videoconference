import React from 'react';

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
        return (<div>
            {!this.state.meetingCreated && (
                <div>
                    Meeting name: <input value={this.state.roomName} onChange={this.handleChangeMeeting} /><br />
                    SSN: <input placeholder="YYYYMMDDXXXX"value={this.state.ssn} onChange={this.handleChange} type="text" /><button disabled={!this.state.buttonEnabled} onClick={this.addSsn}>Add SSN</button>
                    <ul>
                        {this.state.ssns.map(ssn => (<li key={ssn}>{ssn}<button>Remove</button></li>))}
                    </ul>

                    <button onClick={this.addMeeting} disabled={!(this.state.ssns.length > 0 && this.state.roomName.length > 0)}>Create meeting</button>
                </div>)}
            {this.state.meetingCreated && (
                <div>Your meeting room was created. <button onClick={this.clearState}>Create a new meeting!</button></div>
            )}
        </div>);
    };
}