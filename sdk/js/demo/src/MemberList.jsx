import React from "react";

export default class MemberList extends React.Component {
    constructor() {
        super();
        this.state = {
            members: [],
        };
    }

    componentDidMount = () => {
        const { client } = this.props;
        client.on("peer-join", this._handleAddPeer);
        client.on("peer-leave", this._handleRemovePeer);
    };

    componentWillUnmount = () => {
        const { client } = this.props;
        client.off("peer-join", this._handleAddPeer);
        client.off("peer-leave", this._handleRemovePeer);
    };


    _handleAddPeer = async (rid, mid, info) => {
        debugger;
        let members = this.state.members;
        members.push({ mid: mid, info });
        this.setState({ members });
    };

    _handleRemovePeer = async (rid, mid) => {
        let members = this.state.members;
        members = members.filter(item => item.mid !== mid);
        this.setState({ members });
    };

    render() {
        const { members } = this.state;

        return (<ul>
            {members.map(member => (<li style={{color: "white"}}>{member.info.name}</li>))}
        </ul>);
    };
}