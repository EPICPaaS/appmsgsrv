package com.ks.gopush.cli.model.response;

import java.io.Serializable;

public class MsgResponse implements Serializable{
	
	private static final long serialVersionUID = 1L;

	private BaseResponse baseResponse;

	public BaseResponse getBaseResponse() {
		return baseResponse;
	}

	public void setBaseResponse(BaseResponse baseResponse) {
		this.baseResponse = baseResponse;
	}
	

}
